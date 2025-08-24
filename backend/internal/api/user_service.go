package api

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"log"
	"strings"
	"time"

	"github.com/c0re/dft/backend/internal/storage"
)

// UserService - сервис для работы с пользователями
type UserService struct {
	storage storage.Storage
	logger  *log.Logger
}

// NewUserService - создаем новый сервис пользователей
func NewUserService(storage storage.Storage, logger *log.Logger) *UserService {
	return &UserService{
		storage: storage,
		logger:  logger,
	}
}

// RegisterUser - регистрирует нового пользователя
func (s *UserService) RegisterUser(ctx context.Context, req RegisterUserRequest) (*RegisterUserResponse, error) {
	// TODO: Реализовать регистрацию пользователя
	// 1. Валидировать входные данные
	// 2. Проверить что username не занят
	// 3. Сгенерировать ID пользователя
	// 4. Сохранить в storage
	// 5. Вернуть результат

	return nil, errors.New("регистрация пользователя не реализована")
}

// GetUser - получает пользователя по ID
func (s *UserService) GetUser(ctx context.Context, userID string) (*User, error) {
	// TODO: Реализовать получение пользователя
	// 1. Получить из storage по ID
	// 2. Проверить что пользователь существует
	// 3. Вернуть данные (без приватной информации)

	return nil, errors.New("получение пользователя не реализовано")
}

// GetUserByUsername - получает пользователя по username
func (s *UserService) GetUserByUsername(ctx context.Context, username string) (*User, error) {
	// TODO: Реализовать получение по username
	// 1. Получить из storage по username
	// 2. Вернуть данные

	return nil, errors.New("получение пользователя по username не реализовано")
}

// UpdateUserStatus - обновляет статус пользователя
func (s *UserService) UpdateUserStatus(ctx context.Context, userID, status string) error {
	// TODO: Реализовать обновление статуса
	// 1. Проверить валидность статуса
	// 2. Обновить в storage
	// 3. Уведомить других пользователей через P2P

	return errors.New("обновление статуса не реализовано")
}

// SearchUsers - ищет пользователей по username
func (s *UserService) SearchUsers(ctx context.Context, query string, limit int) ([]*User, error) {
	// TODO: Реализовать поиск пользователей
	// 1. Искать в storage по username с LIKE
	// 2. Ограничить количество результатов
	// 3. Вернуть публичные данные

	return nil, errors.New("поиск пользователей не реализован")
}

// ValidateUsername - валидирует username
func (s *UserService) ValidateUsername(username string) error {
	// TODO: Реализовать валидацию
	// 1. Проверить длину (3-50 символов)
	// 2. Проверить символы (только буквы, цифры, подчеркивание)
	// 3. Проверить что не зарезервированное слово

	if len(username) < 3 || len(username) > 50 {
		return errors.New("username должен быть от 3 до 50 символов")
	}

	// Проверяем допустимые символы
	for _, char := range username {
		if !((char >= 'a' && char <= 'z') ||
			 (char >= 'A' && char <= 'Z') ||
			 (char >= '0' && char <= '9') ||
			 char == '_') {
			return errors.New("username может содержать только буквы, цифры и подчеркивание")
		}
	}

	return nil
}

// ValidateDisplayName - валидирует отображаемое имя
func (s *UserService) ValidateDisplayName(displayName string) error {
	// TODO: Реализовать валидацию display name
	// 1. Проверить длину
	// 2. Проверить на запрещенные символы

	if len(displayName) < 1 || len(displayName) > 100 {
		return errors.New("display name должен быть от 1 до 100 символов")
	}

	return nil
}

// ValidatePublicKey - валидирует публичный ключ
func (s *UserService) ValidatePublicKey(publicKey string) error {
	// TODO: Реализовать валидацию публичного ключа
	// 1. Проверить формат (должен быть hex)
	// 2. Проверить длину (для Ed25519)
	// 3. Попробовать распарсить ключ

	if len(publicKey) == 0 {
		return errors.New("публичный ключ не может быть пустым")
	}

	// Проверяем что это валидный hex
	if _, err := hex.DecodeString(publicKey); err != nil {
		return errors.New("публичный ключ должен быть в hex формате")
	}

	return nil
}

// GenerateUserID - генерирует уникальный ID для пользователя
func (s *UserService) GenerateUserID() (string, error) {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return "user_" + hex.EncodeToString(bytes), nil
}

// IsUsernameTaken - проверяет занят ли username
func (s *UserService) IsUsernameTaken(ctx context.Context, username string) (bool, error) {
	// TODO: Реализовать проверку
	// 1. Попробовать получить пользователя по username
	// 2. Если найден - значит занят

	_, err := s.storage.GetUserByUsername(ctx, strings.ToLower(username))
	if err != nil {
		// Если ошибка "не найден" - значит свободен
		return false, nil
	}

	return true, nil
}

// GetUserPublicInfo - получает публичную информацию о пользователе
func (s *UserService) GetUserPublicInfo(ctx context.Context, userID string) (*User, error) {
	// TODO: Реализовать получение публичной информации
	// 1. Получить пользователя из storage
	// 2. Вернуть только публичные поля (без приватных ключей)

	return nil, errors.New("получение публичной информации не реализовано")
}
