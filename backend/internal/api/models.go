package api

import (
	"time"
)

// Message - структура сообщения
type Message struct {
	ID        string    `json:"id" db:"id"`
	FromID    string    `json:"from_id" db:"from_id"`
	ToID      string    `json:"to_id" db:"to_id"`
	Content   string    `json:"content" db:"content"`
	Type      string    `json:"type" db:"type"` // text, image, file, voice
	Status    string    `json:"status" db:"status"` // sent, delivered, read
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`

	// Зашифрованные данные
	EncryptedContent string `json:"-" db:"encrypted_content"`
	Signature        string `json:"-" db:"signature"`
}

// SendMessageRequest - запрос на отправку сообщения
type SendMessageRequest struct {
	ToID    string `json:"to_id" validate:"required"`
	Content string `json:"content" validate:"required"`
	Type    string `json:"type" validate:"required"` // text, image, file, voice
}

// SendMessageResponse - ответ на отправку сообщения
type SendMessageResponse struct {
	MessageID string    `json:"message_id"`
	Status    string    `json:"status"`
	Timestamp time.Time `json:"timestamp"`
}

// GetMessagesRequest - запрос на получение сообщений
type GetMessagesRequest struct {
	WithUserID string `json:"with_user_id"`
	Limit      int    `json:"limit" validate:"max=100"`
	Offset     int    `json:"offset"`
}

// GetMessagesResponse - ответ с сообщениями
type GetMessagesResponse struct {
	Messages []Message `json:"messages"`
	Total    int       `json:"total"`
	Limit    int       `json:"limit"`
	Offset   int       `json:"offset"`
}

// User - структура пользователя
type User struct {
	ID          string    `json:"id" db:"id"`
	Username    string    `json:"username" db:"username"`
	PublicKey   string    `json:"public_key" db:"public_key"`
	DisplayName string    `json:"display_name" db:"display_name"`
	Status      string    `json:"status" db:"status"` // online, offline, away
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

// RegisterUserRequest - запрос на регистрацию
type RegisterUserRequest struct {
	Username    string `json:"username" validate:"required,min=3,max=50"`
	DisplayName string `json:"display_name" validate:"required"`
	PublicKey   string `json:"public_key" validate:"required"`
}

// RegisterUserResponse - ответ на регистрацию
type RegisterUserResponse struct {
	UserID string `json:"user_id"`
	Status string `json:"status"`
}

// APIResponse - общий ответ API
type APIResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

// APIError - структура ошибки API
type APIError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
}
