package main

import (
	"context"
	"fmt"
	"log"

	"github.com/evgeney-fullstack/chat-project-on-go/internal/app"
	"github.com/evgeney-fullstack/chat-project-on-go/internal/config"
	"github.com/evgeney-fullstack/chat-project-on-go/internal/logger_app"
	"github.com/evgeney-fullstack/chat-project-on-go/internal/transport"
)

func main() {
	// Загружаем конфигурацию из переменных окружения (или используем значения по умолчанию).
	cfg, err := config.Load()
	if err != nil { // Если произошла ошибка при загрузке конфигурации.
		log.Fatalf("failed to load config: %v", err) // Выводим ошибку и завершаем программу с кодом 1.
	}

	// Создаём экземпляр нашего логгера, который будет писать в файл, указанный в конфигурации.
	loggerApp, err := logger_app.NewLogger(cfg.LogFile)
	if err != nil { // Если не удалось создать логгер (например, не открывается файл).
		log.Fatalf("failed to init logger: %v", err) // Завершаем программу с ошибкой.
	}

	// Создаём сервис чата, который управляет сессиями клиентов.
	service := app.NewChatService(cfg.MaxClients)

	// Создаём TCP-сервер, передавая адрес, сервис, логгер и максимальное число клиентов.
	server, err := transport.NewTCPServer(cfg.Addr, service, loggerApp, cfg.MaxClients)
	if err != nil { // Если не удалось создать сервер (например, порт уже занят).
		loggerApp.Error("failed to create server", "err", err) // Логируем ошибку через наш логгер.
		return
	}

	// Выводим в консоль сообщение о том, что сервер запущен и на каком адресе слушает.
	fmt.Printf("Server listening on %s\n", cfg.Addr)

	// Создаём контекст с функцией отмены, чтобы иметь возможность остановить сервер по сигналу.
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel() // При выходе из main отменяем контекст (освобождаем ресурсы).

	// Логируем в файл информацию о старте сервера: адрес и максимальное число клиентов.
	loggerApp.Info("server starting", "addr", cfg.Addr, "max_clients", cfg.MaxClients)

	// Запускаем основной цикл сервера: принятие подключений и их обработка.
	// Передаём контекст, чтобы сервер мог корректно завершиться при отмене.
	if err := server.Run(ctx); err != nil && err != context.Canceled {
		loggerApp.Error("server stopped", "err", err)
	}

	// Логируем сообщение о штатном завершении работы сервера.
	loggerApp.Info("server gracefully shutdown")

}
