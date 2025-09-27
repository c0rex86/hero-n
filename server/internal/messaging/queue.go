package messaging

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"
)

type Queue struct {
	db *sql.DB
}

func NewQueue(db *sql.DB) *Queue { return &Queue{db: db} }

func (q *Queue) Enqueue(ctx context.Context, conversationID, messageID string, envelope []byte, sentAt time.Time) error {
	_, err := q.db.ExecContext(ctx, `INSERT OR IGNORE INTO messages (id, conversation_id, message_id, envelope, sent_at, created_at) VALUES (?, ?, ?, ?, ?, ?)`,
		messageID, conversationID, messageID, envelope, sentAt.Unix(), time.Now().Unix())
	if err != nil {
		return fmt.Errorf("enqueue: %w", err)
	}
	return nil
}

func (q *Queue) PullSince(ctx context.Context, conversationID string, sinceUnix int64, limit int) ([][]byte, error) {
	rows, err := q.db.QueryContext(ctx, `SELECT envelope FROM messages WHERE conversation_id = ? AND sent_at >= ? ORDER BY sent_at ASC LIMIT ?`, conversationID, sinceUnix, limit)
	if err != nil {
		return nil, fmt.Errorf("pull: %w", err)
	}
	defer rows.Close()
	var res [][]byte
	for rows.Next() {
		var b []byte
		if err := rows.Scan(&b); err != nil {
			return nil, err
		}
		res = append(res, b)
	}
	return res, nil
}

func (q *Queue) PullPage(ctx context.Context, conversationID string, sinceUnix int64, limit int) (envelopes [][]byte, nextSince int64, hasMore bool, err error) {
	if limit <= 0 { limit = 100 }
	rows, err := q.db.QueryContext(ctx, `SELECT envelope, sent_at FROM messages WHERE conversation_id = ? AND sent_at >= ? ORDER BY sent_at ASC LIMIT ?`, conversationID, sinceUnix, limit+1)
	if err != nil { return nil, 0, false, fmt.Errorf("pull page: %w", err) }
	defer rows.Close()
	var res [][]byte
	var last int64
	for rows.Next() {
		var b []byte
		var ts int64
		if err := rows.Scan(&b, &ts); err != nil { return nil, 0, false, err }
		res = append(res, b)
		last = ts
	}
	if len(res) > limit {
		res = res[:limit]
		return res, last, true, nil
	}
	return res, last, false, nil
}

func (q *Queue) Delete(ctx context.Context, conversationID, messageID string) error {
	res, err := q.db.ExecContext(ctx, `DELETE FROM messages WHERE conversation_id = ? AND message_id = ?`, conversationID, messageID)
	if err != nil {
		return fmt.Errorf("delete: %w", err)
	}
	affected, _ := res.RowsAffected()
	if affected == 0 {
		return errors.New("not found")
	}
	return nil
}
