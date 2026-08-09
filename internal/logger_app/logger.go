package logger_app

import (
	"fmt"
	"io"
	"log"
	"os"
)

// Logger — обёртка над стандартным log.Logger для нашего чата.
type Logger struct {
	logger *log.Logger
}

func NewLogger(path string) (*Logger, error) {
	var out io.Writer

	out, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644) // Открываем (или создаём) файл для записи и добавления в конец, права 0644
	if err != nil {
		return nil, err
	}

	l := log.New(out, "SERVER: ", log.LstdFlags) // Создаём новый стандартный логгер: вывод в out, префикс "SERVER: ", флаги (дата и время)
	return &Logger{logger: l}, nil
}

// Info пишет информационное сообщение .
func (l *Logger) Info(msg string, keysAndValues ...any) {
	args := append([]any{msg}, keysAndValues...) // Создаём срез args: первый элемент — msg, затем все переданные ключи/значения
	l.logger.Println(args...)
}

// Error пишет ошибку с префиксом ERROR:
func (l *Logger) Error(msg string, keysAndValues ...any) {
	args := append([]any{msg}, keysAndValues...)
	formatted := formatArgs(args...)
	l.logger.Printf("ERROR: %s\n", formatted)
}

// formatArgs преобразует набор аргументов в строку вида "сообщение | ключ1=значение1 ключ2=значение2 ..."
func formatArgs(args ...any) string {
	if len(args) == 0 {
		return ""
	}

	if len(args) == 1 {
		return fmtToString(args[0])
	}

	result := fmtToString(args[0]) + " | " // Начинаем результат: преобразуем первый аргумент (сообщение) и добавляем разделитель " | "

	for i := 1; i < len(args); i += 2 {
		key, ok := args[i].(string) // Пытаемся привести текущий аргумент к строке (ключ должен быть строкой)
		if !ok {
			continue
		}
		var val string
		if i+1 < len(args) {
			val = fmtToString(args[i+1])
		} else {
			val = "<nil>"
		}
		result += key + "=" + val + " "
	}
	return result
}

// fmtToString преобразует любое значение v в строку (учитывая типы string, error и все остальные)
func fmtToString(v any) string {
	//определяим тип значения
	switch t := v.(type) {
	case string:
		return t
	case error:
		return t.Error()
	default:
		return fmt.Sprintf("%v", t)
	}
}
