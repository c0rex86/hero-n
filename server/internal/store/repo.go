package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"dev.c0rex64.heroin/internal/auth"
)

type AuthRepo struct { db *sql.DB }

func NewAuthRepo(db *sql.DB) *AuthRepo { return &AuthRepo{db: db} }

func (r *AuthRepo) CreateUser(ctx context.Context, u auth.UserRecord) error {
	_, err := r.db.ExecContext(ctx, `INSERT INTO users (id, username, server_salt, password_hash, public_key, second_factor_secret, created_at) VALUES (?, ?, ?, ?, ?, ?, ?)`,
		u.ID, u.Username, u.ServerSalt, u.PasswordHash, u.PublicKey, u.SecondFactorSecret, u.CreatedAt)
	if err != nil { return fmt.Errorf("create user: %w", err) }
	return nil
}

func (r *AuthRepo) FindUserByUsername(ctx context.Context, username string) (*auth.UserRecord, error) {
	row := r.db.QueryRowContext(ctx, `SELECT id, username, server_salt, password_hash, public_key, second_factor_secret, created_at FROM users WHERE username = ?`, username)
	var u auth.UserRecord
	if err := row.Scan(&u.ID, &u.Username, &u.ServerSalt, &u.PasswordHash, &u.PublicKey, &u.SecondFactorSecret, &u.CreatedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) { return nil, nil }
		return nil, fmt.Errorf("find user: %w", err)
	}
	return &u, nil
}

func (r *AuthRepo) GetPublicKey(ctx context.Context, userID string) ([]byte, error) {
	row := r.db.QueryRowContext(ctx, `SELECT public_key FROM users WHERE id = ?`, userID)
	var pk []byte
	if err := row.Scan(&pk); err != nil {
		return nil, fmt.Errorf("get public key: %w", err)
	}
	return pk, nil
}

func (r *AuthRepo) CreateSession(ctx context.Context, userID string, deviceID string, refreshHash []byte, expiresAt time.Time) (string, error) {
	id := generateUUIDv7()
	_, err := r.db.ExecContext(ctx, `INSERT INTO sessions (id, user_id, device_id, refresh_token_hash, expires_at, created_at) VALUES (?, ?, ?, ?, ?, ?)`,
		id, userID, deviceID, refreshHash, expiresAt.Unix(), time.Now().Unix())
	if err != nil { return "", fmt.Errorf("create session: %w", err) }
	return id, nil
}

func (r *AuthRepo) GetSessionByRefreshHash(ctx context.Context, refreshHash []byte) (string, string, time.Time, error) {
	row := r.db.QueryRowContext(ctx, `SELECT user_id, device_id, expires_at FROM sessions WHERE refresh_token_hash = ?`, refreshHash)
	var userID, deviceID string
	var expUnix int64
	if err := row.Scan(&userID, &deviceID, &expUnix); err != nil {
		return "", "", time.Time{}, fmt.Errorf("get session: %w", err)
	}
	return userID, deviceID, time.Unix(expUnix, 0), nil
}

func (r *AuthRepo) WriteAudit(ctx context.Context, userID, deviceID, eventType string) error {
	id := generateUUIDv7()
	_, err := r.db.ExecContext(ctx, `INSERT INTO audit_logs (id, user_id, device_id, event_type, created_at) VALUES (?, ?, ?, ?, ?)`,
		id, userID, deviceID, eventType, time.Now().Unix())
	if err != nil { return fmt.Errorf("write audit: %w", err) }
	return nil
}

func generateUUIDv7() string {
	now := time.Now().UnixMilli()
	b := make([]byte, 16)
	for i := 0; i < 6; i++ { b[i] = byte(now >> (8 * (5 - i))) }
	b[6] = 0x70 | 0x0F&b[6]
	b[8] = 0x80 | 0x3F&b[8]
	for i := 9; i < 16; i++ { b[i] = byte(time.Now().UnixNano() >> (8 * (i % 8))) }
	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:16])
}
