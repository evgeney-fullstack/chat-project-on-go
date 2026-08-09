package app

import (
	"errors"
	"sync"
)

// Глобальные переменные-ошибки, которые могут возвращаться методами сервиса
var (
	ErrClientNotFound = errors.New("client not found")                  // Клиент с таким ID не найден
	ErrMaxClients     = errors.New("maximum number of clients reached") // Превышено максимальное количество клиентов
	ErrClientBusy     = errors.New("client busy (send channel full)")   // Клиент занят (его канал отправки заполнен)
)

type ClientID string

// Session представляет активное подключение клиента
type Session struct {
	ID       ClientID    // Уникальный идентификатор клиента
	SendChan chan []byte // Буферизированный канал для отправки сообщений этому клиенту
}

// ChatService — основной объект, хранящий все активные сессии и управляющий ими
type ChatService struct {
	mu         sync.Mutex            // Мьютекс для защиты доступа к карте sessions
	sessions   map[ClientID]*Session // Карта: ключ — ID клиента, значение — указатель на его сессию
	maxClients int                   // Максимальное допустимое количество одновременных клиентов
}

// NewChatService создаёт новый экземпляр ChatService с заданным максимальным числом клиентов
func NewChatService(maxClients int) *ChatService {
	return &ChatService{
		sessions:   make(map[ClientID]*Session), // Инициализируем пустую карту
		maxClients: maxClients,                  // Запоминаем лимит
	}
}

// AddSession добавляет нового клиента в сервис (создаёт сессию и канал)
func (s *ChatService) AddSession(id ClientID) (*Session, error) {
	s.mu.Lock()
	defer s.mu.Unlock() // Гарантируем разблокировку при выходе из функции

	// Проверяем, не превышен ли лимит клиентов
	if len(s.sessions) >= s.maxClients {
		return nil, ErrMaxClients
	}

	// Создаём новую сессию с буферизированным каналом размером 16 сообщений
	sess := &Session{
		ID:       id,
		SendChan: make(chan []byte, 16),
	}
	// Сохраняем сессию в карте по ID клиента
	s.sessions[id] = sess
	return sess, nil // Возвращаем созданную сессию и nil-ошибку
}

// RemoveSession удаляет клиента из сервиса по ID
func (s *ChatService) RemoveSession(id ClientID) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.sessions, id) // Удаляем запись из карты (канал остаётся открытым, но на него больше нет ссылок из карты)

}

// GetSessions возвращает срез указателей на все текущие сессии
func (s *ChatService) GetSessions() []*Session {
	s.mu.Lock()
	defer s.mu.Unlock()

	res := make([]*Session, 0, len(s.sessions))
	// Проходим по всем значениям карты и добавляем их в результат
	for _, v := range s.sessions {
		res = append(res, v)
	}
	return res
}

// Forward пересылает сообщение от одного клиента другому по их ID
func (s *ChatService) Forward(fromID, toID ClientID, msg []byte) error {
	s.mu.Lock()
	to, ok := s.sessions[toID] // Ищем получателя в карте
	s.mu.Unlock()

	if !ok {
		return ErrClientNotFound // Возвращаем соответствующую ошибку
	}

	// Пытаемся отправить сообщение в канал получателя без блокировки (неблокирующий select)
	select {
	case to.SendChan <- msg: // Если канал готов принять сообщение (есть свободное место)
		return nil // Успешно отправлено
	default: // Если канал заполнен (или закрыт)
		return ErrClientBusy // Возвращаем ошибку "клиент занят"
	}
}

// Shutdown корректно завершает работу сервиса: закрывает каналы всех сессий и очищает карту
func (s *ChatService) Shutdown() {
	s.mu.Lock()         // Блокируем мьютекс на всё время
	defer s.mu.Unlock() // Разблокируем при выходе

	// Идём по всем сессиям
	for _, sess := range s.sessions {
		close(sess.SendChan) // Закрываем канал отправки каждого клиента (сигнал для читающей горутины о завершении)
	}
	// Пересоздаём карту, удаляя все ссылки на старые сессии (они будут собраны сборщиком мусора)
	s.sessions = make(map[ClientID]*Session)
}
