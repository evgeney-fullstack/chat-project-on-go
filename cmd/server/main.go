package main

import (
	"log"

	"github.com/evgeney-fullstack/chat-project-on-go/internal/app"
	"github.com/evgeney-fullstack/chat-project-on-go/internal/config"
	"github.com/evgeney-fullstack/chat-project-on-go/internal/logger_app"
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

}
