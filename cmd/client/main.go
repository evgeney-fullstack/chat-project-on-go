package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/evgeney-fullstack/chat-project-on-go/internal/config" // Пакет для загрузки конфигурации из .env
	"github.com/evgeney-fullstack/chat-project-on-go/internal/transport"
	// Пакет с реализацией TCP-клиента и сервера
)

func main() {
	// Загружаем конфигурацию  из переменных окружения
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	// Создаём TCP-клиент и подключаемся к серверу по адресу из конфига
	client, err := transport.NewTCPClient(cfg.Addr)
	if err != nil {
		log.Fatalf("failed to connect: %v", err)
	}
	defer client.Close() // Гарантируем закрытие соединения при выходе из main

	// Создаём контекст с функцией отмены (для graceful shutdown)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Запускаем горутину для перехвата системных сигналов
	go func() {
		sigChan := make(chan os.Signal, 1)                      // Буферизированный канал для сигналов (буфер 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM) // Подписываемся на Ctrl+C (SIGINT) и завершение (SIGTERM)
		<-sigChan
		cancel()
	}()

	// Запускаем основной цикл клиента
	if err := client.Run(ctx); err != nil && err != context.Canceled {
		log.Printf("client stopped: %v", err)
	}
}
