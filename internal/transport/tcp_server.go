package transport

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"sync"

	"github.com/evgeney-fullstack/chat-project-on-go/internal/app"
	"github.com/evgeney-fullstack/chat-project-on-go/internal/logger_app"
)

type TCPServer struct {
	listener   net.Listener
	service    *app.ChatService
	logger     *logger_app.Logger
	mu         sync.Mutex
	clients    int
	maxClients int
	cancel     context.CancelFunc // для остановки сервера по команде SHUTDOWN
}

func NewTCPServer(addr string, service *app.ChatService, logger *logger_app.Logger, maxClients int) (*TCPServer, error) {
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, err
	}
	return &TCPServer{
		listener:   ln,
		service:    service,
		logger:     logger,
		maxClients: maxClients,
	}, nil
}

func (s *TCPServer) Run(ctx context.Context) error {
	ctx, cancel := context.WithCancel(ctx)
	s.cancel = cancel
	defer cancel()

	// Горутина, которая закрывает слушатель при отмене контекста (например, по SHUTDOWN)
	go func() {
		<-ctx.Done()
		s.listener.Close()
	}()

	for {
		conn, err := s.listener.Accept()
		if err != nil {
			select {
			case <-ctx.Done():
				return ctx.Err() // нормальное завершение, если слушатель закрыт из-за отмены
			default:
				s.logger.Error("accept error", "err", err)
				continue
			}
		}

		// Проверка лимита клиентов до принятия соединения
		s.mu.Lock()
		if s.clients >= s.maxClients {
			conn.Write([]byte(MsgMaxClients + "\n"))
			conn.Close()
			s.mu.Unlock()
			continue
		}
		s.clients++
		s.mu.Unlock()

		go s.handleConnection(conn)
	}
}

func (s *TCPServer) handleConnection(conn net.Conn) {
	defer func() {
		conn.Close()
		s.mu.Lock()
		s.clients--
		s.mu.Unlock()
		s.logger.Info("client disconnected", "id", conn.RemoteAddr().String())
	}()

	id := app.ClientID(conn.RemoteAddr().String())

	sess, err := s.service.AddSession(id)
	if err != nil {
		s.logger.Error("failed to add session", "id", string(id), "err", err)
		conn.Write([]byte("ERROR: server internal error\n"))
		return
	}
	defer s.service.RemoveSession(id)

	switch s.clients {
	case 1:
		fmt.Println("Client #1 connected")
	case 2:
		fmt.Println("Client #2 connected")
		fmt.Println("Both clients connected. Chat started!")
		// Рассылаем всем, что чат начался
		for _, other := range s.service.GetSessions() {
			select {
			case other.SendChan <- []byte(MsgChatStarted + "\n"):
			default:
			}
		}
	}

	// Горутина отправки: читает из канала сессии и пишет в сокет
	go func() {
		for msg := range sess.SendChan {
			_, err := conn.Write(msg)
			if err != nil {
				s.logger.Error("write error", "id", string(id), "err", err)
				return
			}
		}
	}()

	reader := bufio.NewReader(conn)
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			if errors.Is(err, io.EOF) {
				s.logger.Info("client read EOF", "id", string(id))
			} else {
				s.logger.Error("read error", "id", string(id), "err", err)
			}
			break
		}

		msg, ok := ParseLine([]byte(line))
		if !ok {
			continue
		}

		// Обработка системной команды SHUTDOWN (только от одного клиента)
		if IsCommand(msg) {
			if msg == CmdShutdown {
				s.logger.Info("SHUTDOWN command received", "id", string(id))
				// Уведомляем остальных клиентов о завершении работы
				for _, other := range s.service.GetSessions() {
					if other.ID != id {
						select {
						case other.SendChan <- []byte(MsgServerShutdown + "\n"):
						default:
						}
					}
				}
				s.service.Shutdown() // закрываем все каналы сессий
				// Останавливаем сервер через отмену контекста
				if s.cancel != nil {
					s.cancel()
				}
				return
			}
			continue
		}

		// Рассылка обычного сообщения всем, кроме отправителя
		sessions := s.service.GetSessions()
		for _, other := range sessions {
			if other.ID == id {
				continue
			}
			payload := []byte(msg + "\n")
			if fErr := s.service.Forward(id, other.ID, payload); fErr != nil {
				s.logger.Error("failed to forward message", "to", string(other.ID), "err", fErr)
			}
		}
	}

	// При отключении клиента уведомляем остальных о том, что собеседник вышел
	for _, other := range s.service.GetSessions() {
		if other.ID != id {
			select {
			case other.SendChan <- []byte(MsgOtherDisconnected + "\n"):
			default:
			}
		}
	}
}
