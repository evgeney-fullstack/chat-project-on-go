package config

import (
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv" // Сторонний пакет для загрузки .env-файлов
)

// Определяем структуру Config, которая будет хранить все параметры конфигурации
type Config struct {
	Addr       string // Адрес (хост + порт), на котором будет работать сервер
	MaxClients int    // Максимальное количество одновременных клиентов
	LogFile    string // Путь к файлу для записи логов
}

// Функция Load загружает конфигурацию из файла .env и возвращает указатель на Config и ошибку
func Load() (*Config, error) {

	// Загружаем переменные окружения из файла config.env (по умолчанию ищет в текущей директории)
	if err := godotenv.Load("config.env"); err != nil {
		// Если файл не найден или не может быть прочитан, логируем фатальную ошибку и завершаем программу
		log.Fatalf("error loading env variables: %s", err.Error())
	}

	// Читаем переменную CHAT_ADDR и CHAT_PORT, объединяем их через двоеточие, получая полный адрес
	addr := os.Getenv("CHAT_ADDR") + ":" + os.Getenv("CHAT_PORT")

	// Читаем переменную MAX_CLIENTS (строка) и преобразуем её в целое число
	maxClients, err := strconv.Atoi(os.Getenv("MAX_CLIENTS"))
	if err != nil {
		// Если строка не является числом или пуста, вызываем фатальную ошибку и завершаем программу
		log.Fatalf("error value number of clients is incorrect: %s", err.Error())
	}

	// Читаем переменную CHAT_LOG_FILE – путь к файлу логов
	logFile := os.Getenv("CHAT_LOG_FILE")

	// Возвращаем указатель на структуру Config с заполненными полями и ошибку nil
	return &Config{
		Addr:       addr,
		MaxClients: maxClients,
		LogFile:    logFile,
	}, nil
}
