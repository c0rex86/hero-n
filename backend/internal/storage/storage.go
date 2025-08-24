package storage

import (
	"context"
	"database/sql"
	"log"

	_ "github.com/lib/pq" // PostgreSQL драйвер
)

// Storage - интерфейс для хранилища данных
type Storage interface {
	// Сообщения
	SaveMessage(ctx context.Context, msg *DBMessage) error
	GetMessage(ctx context.Context, id string) (*DBMessage, error)
	GetMessagesBetweenUsers(ctx context.Context, user1ID, user2ID string, limit, offset int) ([]*DBMessage, error)
	GetRecentMessages(ctx context.Context, userID string, limit int) ([]*DBMessage, error)

	// Пользователи
	SaveUser(ctx context.Context, user *DBUser) error
	GetUser(ctx context.Context, id string) (*DBUser, error)
	GetUserByUsername(ctx context.Context, username string) (*DBUser, error)
	UpdateUserStatus(ctx context.Context, userID, status string) error

	// Сессии
	SaveSession(ctx context.Context, session *DBUserSession) error
	GetSession(ctx context.Context, token string) (*DBUserSession, error)
	DeleteExpiredSessions(ctx context.Context) error

	// Системные
	InitializeDatabase(ctx context.Context) error
	Close() error
}

// PostgresStorage - реализация хранилища на PostgreSQL
type PostgresStorage struct {
	db     *sql.DB
	logger *log.Logger
}

// NewStorage - создаем новое хранилище
func NewStorage(ctx context.Context, logger *log.Logger) (Storage, error) {
	// TODO: Настроить подключение к PostgreSQL
	// 1. Прочитать конфигурацию из переменных окружения
	// 2. Подключиться к БД с retry логикой
	// 3. Инициализировать таблицы

	storage := &PostgresStorage{
		logger: logger,
	}

	// Инициализируем базу данных
	if err := storage.InitializeDatabase(ctx); err != nil {
		return nil, err
	}

	logger.Println("Хранилище PostgreSQL запущено")

	return storage, nil
}

// InitializeDatabase - инициализация базы данных
func (s *PostgresStorage) InitializeDatabase(ctx context.Context) error {
	// TODO: Реализовать инициализацию БД
	// 1. Подключиться к PostgreSQL
	// 2. Выполнить SQL скрипты создания таблиц
	// 3. Создать индексы и триггеры
	// 4. Проверить подключение

	return nil
}

// SaveMessage - сохраняет сообщение
func (s *PostgresStorage) SaveMessage(ctx context.Context, msg *DBMessage) error {
	// TODO: Реализовать сохранение сообщения
	// 1. Начать транзакцию
	// 2. Вставить сообщение в таблицу
	// 3. Обновить статусы пользователей
	// 4. Зафиксировать транзакцию

	return nil
}

// GetMessage - получает сообщение по ID
func (s *PostgresStorage) GetMessage(ctx context.Context, id string) (*DBMessage, error) {
	// TODO: Реализовать получение сообщения
	// 1. Выполнить SELECT запрос
	// 2. Сканировать результат в структуру
	// 3. Проверить что сообщение существует

	return nil, nil
}

// GetMessagesBetweenUsers - получает сообщения между двумя пользователями
func (s *PostgresStorage) GetMessagesBetweenUsers(ctx context.Context, user1ID, user2ID string, limit, offset int) ([]*DBMessage, error) {
	// TODO: Реализовать получение диалога
	// 1. Использовать индекс conversation
	// 2. Применить пагинацию
	// 3. Отсортировать по времени

	return nil, nil
}

// GetRecentMessages - получает последние сообщения пользователя
func (s *PostgresStorage) GetRecentMessages(ctx context.Context, userID string, limit int) ([]*DBMessage, error) {
	// TODO: Реализовать получение последних сообщений
	// 1. Использовать GetRecentMessagesQuery
	// 2. Группировать по диалогам
	// 3. Ограничить количество

	return nil, nil
}

// SaveUser - сохраняет пользователя
func (s *PostgresStorage) SaveUser(ctx context.Context, user *DBUser) error {
	// TODO: Реализовать сохранение пользователя
	// 1. Проверить уникальность username
	// 2. Вставить в таблицу users
	// 3. Обработать ошибку дублирования

	return nil
}

// GetUser - получает пользователя по ID
func (s *PostgresStorage) GetUser(ctx context.Context, id string) (*DBUser, error) {
	// TODO: Реализовать получение пользователя
	// 1. Выполнить SELECT по ID
	// 2. Сканировать в структуру

	return nil, nil
}

// GetUserByUsername - получает пользователя по username
func (s *PostgresStorage) GetUserByUsername(ctx context.Context, username string) (*DBUser, error) {
	// TODO: Реализовать получение по username
	// 1. Выполнить SELECT по username
	// 2. Использовать индекс для быстрого поиска

	return nil, nil
}

// UpdateUserStatus - обновляет статус пользователя
func (s *PostgresStorage) UpdateUserStatus(ctx context.Context, userID, status string) error {
	// TODO: Реализовать обновление статуса
	// 1. Обновить поле status
	// 2. Установить last_seen

	return nil
}

// SaveSession - сохраняет сессию пользователя
func (s *PostgresStorage) SaveSession(ctx context.Context, session *DBUserSession) error {
	// TODO: Реализовать сохранение сессии
	// 1. Вставить в user_sessions
	// 2. Проверить уникальность

	return nil
}

// GetSession - получает сессию по токену
func (s *PostgresStorage) GetSession(ctx context.Context, token string) (*DBUserSession, error) {
	// TODO: Реализовать получение сессии
	// 1. Найти по токену
	// 2. Проверить срок действия

	return nil, nil
}

// DeleteExpiredSessions - удаляет просроченные сессии
func (s *PostgresStorage) DeleteExpiredSessions(ctx context.Context) error {
	// TODO: Реализовать очистку сессий
	// 1. Удалить записи где expires_at < now
	// 2. Логировать количество удаленных

	return nil
}

// Close - закрывает соединение с базой данных
func (s *PostgresStorage) Close() error {
	s.logger.Println("Закрываем хранилище")

	if s.db != nil {
		return s.db.Close()
	}

	return nil
}
