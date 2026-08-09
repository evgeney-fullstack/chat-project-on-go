package main

import (
	"log"

	"github.com/evgeney-fullstack/chat-project-on-go/internal/config"
)

func main() {
	// Загружаем конфигурацию из переменных окружения (или используем значения по умолчанию).
	cfg, err := config.Load()
	if err != nil { // Если произошла ошибка при загрузке конфигурации.
		log.Fatalf("failed to load config: %v", err) // Выводим ошибку и завершаем программу с кодом 1.
	}

}
