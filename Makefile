# Директория для бинарников
BUILD_DIR = build

# Исполняемые файлы
CLIENT_BIN = $(BUILD_DIR)/client
SERVER_BIN = $(BUILD_DIR)/server

# Конфигурационный файл
CONFIG_SRC = config.env
CONFIG_DST = $(BUILD_DIR)/config.env

# Команда Go
GOFLAGS = -trimpath

# Цели по умолчанию
.PHONY: all clean client server copy-config

all: copy-config client server 

# Копирование или создание config.env в папке build
copy-config:
	mkdir -p $(BUILD_DIR)
	@if [ -f $(CONFIG_SRC) ]; then \
		cp $(CONFIG_SRC) $(CONFIG_DST); \
		echo "Copied $(CONFIG_SRC) to $(CONFIG_DST)"; \
	else \
		echo "# Configuration" > $(CONFIG_DST); \
		echo "CHAT_ADDR = localhost" >> $(CONFIG_DST); \
		echo "CHAT_PORT = 8080" >> $(CONFIG_DST); \
		echo "MAX_CLIENTS = 2" >> $(CONFIG_DST); \
		echo "CHAT_LOG_FILE = server.log" >> $(CONFIG_DST); \
		echo "Created default $(CONFIG_DST)"; \
	fi

# Сборка клиента
client: copy-config
	mkdir -p $(BUILD_DIR)
	go build $(GOFLAGS) -o $(CLIENT_BIN) ./cmd/client

# Сборка сервера
server: copy-config
	mkdir -p $(BUILD_DIR)
	go build $(GOFLAGS) -o $(SERVER_BIN) ./cmd/server



# Очистка
clean:
	rm -rf $(BUILD_DIR)