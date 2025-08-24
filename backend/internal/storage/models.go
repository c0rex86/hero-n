package storage

import (
	"database/sql"
	"time"
)

// DBMessage - структура сообщения в БД
type DBMessage struct {
	ID               string       `db:"id"`
	FromID           string       `db:"from_id"`
	ToID             string       `db:"to_id"`
	Content          string       `db:"content"`          // зашифрованное содержимое
	EncryptedContent string       `db:"encrypted_content"` // дополнительное шифрование
	Signature        string       `db:"signature"`        // подпись отправителя
	Type             string       `db:"type"`             // text, image, file, voice
	Status           string       `db:"status"`           // sent, delivered, read
	CreatedAt        time.Time    `db:"created_at"`
	UpdatedAt        time.Time    `db:"updated_at"`
	DeletedAt        sql.NullTime `db:"deleted_at"`       // мягкое удаление
}

// DBUser - структура пользователя в БД
type DBUser struct {
	ID          string    `db:"id"`
	Username    string    `db:"username"`
	PublicKey   string    `db:"public_key"`
	DisplayName string    `db:"display_name"`
	Status      string    `db:"status"`      // online, offline, away
	LastSeen    time.Time `db:"last_seen"`
	CreatedAt   time.Time `db:"created_at"`
	UpdatedAt   time.Time `db:"updated_at"`
}

// DBUserSession - структура сессии пользователя
type DBUserSession struct {
	ID        string    `db:"id"`
	UserID    string    `db:"user_id"`
	Token     string    `db:"token"`
	IPAddress string    `db:"ip_address"`
	UserAgent string    `db:"user_agent"`
	ExpiresAt time.Time `db:"expires_at"`
	CreatedAt time.Time `db:"created_at"`
}

// DBContact - структура контакта пользователя
type DBContact struct {
	ID         string    `db:"id"`
	UserID     string    `db:"user_id"`
	ContactID  string    `db:"contact_id"`
	Nickname   string    `db:"nickname"`
	AddedAt    time.Time `db:"added_at"`
	Blocked    bool      `db:"blocked"`
}

// MessageQuery - параметры для запроса сообщений
type MessageQuery struct {
	UserID   string
	WithUser string // ID собеседника
	Limit    int
	Offset   int
	Before   time.Time // сообщения до этой даты
	After    time.Time // сообщения после этой даты
}

// UserQuery - параметры для запроса пользователей
type UserQuery struct {
	Username   string
	Limit      int
	Offset     int
	OnlineOnly bool
}

// InsertMessageParams - параметры для вставки сообщения
type InsertMessageParams struct {
	ID               string
	FromID           string
	ToID             string
	Content          string
	EncryptedContent string
	Signature        string
	Type             string
	Status           string
}

// InsertUserParams - параметры для вставки пользователя
type InsertUserParams struct {
	ID          string
	Username    string
	PublicKey   string
	DisplayName string
	Status      string
}

// UpdateUserStatusParams - параметры для обновления статуса
type UpdateUserStatusParams struct {
	UserID  string
	Status  string
	LastSeen time.Time
}
