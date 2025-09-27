package messaging

import (
	"context"
	"encoding/json"
	"time"
)

type Service struct {
	q *Queue
}

type EnvelopeData struct {
	ConversationID string `json:"conversation_id"`
	MessageID      string `json:"message_id"`
	Ciphertext     []byte `json:"ciphertext"`
	Signature      []byte `json:"signature"`
	SentAtUnix     int64  `json:"sent_at_unix"`
}

func NewService(q *Queue) *Service { return &Service{q: q} }

func (s *Service) Send(ctx context.Context, envelope []byte) error {
	var env EnvelopeData
	if err := json.Unmarshal(envelope, &env); err != nil {
		// fallback for raw envelope
		return s.q.Enqueue(ctx, "default", stringHash(envelope), envelope, time.Now())
	}
	sentAt := time.Unix(env.SentAtUnix, 0)
	return s.q.Enqueue(ctx, env.ConversationID, env.MessageID, envelope, sentAt)
}

func (s *Service) Pull(ctx context.Context, conversationID string, since int64) ([][]byte, error) {
	return s.q.PullSince(ctx, conversationID, since, 100)
}

func stringHash(b []byte) string {
	var h uint64
	for _, v := range b { h = h*131 + uint64(v) }
	return toHex(h)
}

func toHex(v uint64) string {
	const hexdigits = "0123456789abcdef"
	buf := make([]byte, 16)
	for i := 15; i >= 0; i-- { buf[i] = hexdigits[v&0xf]; v >>= 4 }
	return string(buf)
}
