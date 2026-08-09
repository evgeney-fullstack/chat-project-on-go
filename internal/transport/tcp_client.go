package transport

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"net"
	"os"
	"sync/atomic"
	"time"
)

// TCPClient — клиент для подключения к чат-серверу по TCP
type TCPClient struct {
	conn          net.Conn // Соединение с сервером
	reader        *bufio.Reader
	writer        *bufio.Writer
	stats         ClientStats   // Статистика соединения
	closeCh       chan struct{} // Канал для сигнала закрытия (закрывается при вызове Close)
	closed        bool
	chatStartedCh chan struct{} // Канал, который закрывается при получении MsgChatStarted от сервера
}

// ClientStats хранит статистику соединения
type ClientStats struct {
	connectTime time.Time // Время подключения
	sent        int64     // Количество отправленных сообщений
	received    int64     // Количество полученных сообщений
}

// NewTCPClient создаёт новое соединение с сервером по адресу addr
func NewTCPClient(addr string) (*TCPClient, error) {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return nil, err
	}
	fmt.Printf("Connected to %s\n", conn.RemoteAddr().String())
	return &TCPClient{
		conn:          conn,
		reader:        bufio.NewReader(conn),
		writer:        bufio.NewWriter(conn),
		stats:         ClientStats{connectTime: time.Now()},
		closeCh:       make(chan struct{}),
		chatStartedCh: make(chan struct{}),
	}, nil
}

// Run запускает клиент: стартует горутины чтения и ввода, ждёт завершения
func (c *TCPClient) Run(ctx context.Context) error {
	go c.readLoop(ctx)  // Читает сообщения из сокета
	go c.inputLoop(ctx) // Читает ввод с клавиатуры

	select {
	case <-ctx.Done(): // Если контекст отменён (например, сигнал Ctrl+C)
		c.Close()
		return ctx.Err()
	case <-c.closeCh: // Если клиент закрыт изнутри (команда EXIT или ошибка)
		return nil
	}
}

// readLoop читает строки из сокета и обрабатывает их
func (c *TCPClient) readLoop(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}
		line, err := c.reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				fmt.Println("Server closed the connection.")
			} else {
				fmt.Fprintf(os.Stderr, "read error: %v\n", err)
			}
			c.Close() // При ошибке чтения закрываем клиент
			return
		}
		msg, ok := ParseLine([]byte(line))
		if !ok {
			continue
		}

		// Обработка служебных сообщений от сервера
		switch msg {
		case MsgChatStarted:
			fmt.Println(msg)
			// Закрываем канал, чтобы inputLoop начал принимать ввод
			select {
			case <-c.chatStartedCh:
				// уже закрыт
			default:
				close(c.chatStartedCh)
			}
		case MsgServerShutdown, MsgMaxClients:
			fmt.Printf("\n%s", msg) // Печатаем без дополнительных префиксов

		case MsgOtherDisconnected:
			fmt.Printf("\n%s\nYou can exit the chat using the [EXIT] command or wait for a new friend:\nYou: ", msg)

		default:
			// Обычное сообщение от собеседника
			fmt.Printf("\nFriend: %s \nYou: ", msg)
		}
		atomic.AddInt64(&c.stats.received, 1) // Атомарно увеличиваем счётчик полученных
	}
}

// inputLoop читает ввод пользователя и отправляет на сервер
func (c *TCPClient) inputLoop(ctx context.Context) {
	// Ждём, пока не получим сигнал о начале чата (MsgChatStarted) или отмену контекста
	select {
	case <-c.chatStartedCh:
		// чат начался – продолжаем
	case <-ctx.Done():
		return
	}

	scanner := bufio.NewScanner(os.Stdin)
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		fmt.Print("You: ")
		if !scanner.Scan() {
			if err := scanner.Err(); err != nil {
				fmt.Fprintf(os.Stderr, "input error: %v\n", err)
			}
			c.Close()
			return
		}
		text := scanner.Text()
		if text == "" {
			continue
		}

		// Обработка локальных команд клиента
		if IsCommand(text) {
			switch text {
			case CmdExit:
				c.Close() // Команда EXIT – закрываем соединение
				return
			case CmdStats:
				c.printStats() // Выводим статистику
				continue
			default:
				// Неизвестные команды отправляем как обычный текст (сервер их проигнорирует)
			}
		}

		// Отправляем введённую строку на сервер
		_, err := c.writer.WriteString(text + "\n")
		if err != nil {
			fmt.Fprintf(os.Stderr, "write error: %v\n", err)
			c.Close()
			return
		}
		if err := c.writer.Flush(); err != nil {
			fmt.Fprintf(os.Stderr, "flush error: %v\n", err)
			c.Close()
			return
		}
		atomic.AddInt64(&c.stats.sent, 1) // Атомарно увеличиваем счётчик отправленных
	}
}

// printStats выводит статистику соединения (время работы, количество сообщений)
func (c *TCPClient) printStats() {
	uptime := time.Since(c.stats.connectTime)
	fmt.Printf("STATS: connected since %v, sent %d, received %d\n",
		uptime.Round(time.Second),
		atomic.LoadInt64(&c.stats.sent),
		atomic.LoadInt64(&c.stats.received))
}

// Close закрывает соединение и сигнализирует о завершении
func (c *TCPClient) Close() {
	if c.closed {
		return // Предотвращаем повторное закрытие
	}
	c.closed = true
	close(c.closeCh) // Сигнал для Run о завершении
	c.conn.Close()
}
