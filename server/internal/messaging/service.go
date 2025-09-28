package messaging

import (
    "context"
    "crypto/ed25519"
    "encoding/binary"
    "encoding/json"
    "errors"
    "time"

    "dev.c0rex64.heroin/internal/crypto"
)

type PublicKeyProvider interface {
	GetPublicKey(ctx context.Context, userID string) ([]byte, error)
}

type Service struct {
	q       *Queue
	kp      PublicKeyProvider
	ratchet *crypto.DoubleRatchet
}

type EnvelopeData struct {
	ConversationID string `json:"conversation_id"`
	MessageID      string `json:"message_id"`
	Ciphertext     []byte `json:"ciphertext"`
	Signature      []byte `json:"signature"`
	SentAtUnix     int64  `json:"sent_at_unix"`
	SenderID       string `json:"sender_id"`
}

func NewService(q *Queue, kp PublicKeyProvider) *Service {
	return &Service{q: q, kp: kp, ratchet: &crypto.DoubleRatchet{}}
}

func (s *Service) Send(ctx context.Context, envelope []byte) error {
	var env EnvelopeData
	if err := json.Unmarshal(envelope, &env); err != nil {
		return err
	}
	if env.ConversationID == "" || env.MessageID == "" { return errors.New("missing ids") }
	if len(env.Signature) == 0 { return errors.New("missing signature") }
	if env.SenderID == "" { return errors.New("missing sender") }
	pk, err := s.kp.GetPublicKey(ctx, env.SenderID)
	if err != nil { return err }
	if len(pk) != ed25519.PublicKeySize { return errors.New("invalid public key") }
	payload := signPayload(env)
	if !ed25519.Verify(ed25519.PublicKey(pk), payload, env.Signature) {
		return errors.New("bad signature")
	}
	sentAt := time.Unix(env.SentAtUnix, 0)
	return s.q.Enqueue(ctx, env.ConversationID, env.MessageID, envelope, sentAt)
}

func (s *Service) Pull(ctx context.Context, conversationID string, since int64) ([][]byte, error) {
	return s.q.PullSince(ctx, conversationID, since, 100)
}

func (s *Service) PullPage(ctx context.Context, conversationID string, since int64, limit int) ([][]byte, int64, bool, error) {
	return s.q.PullPage(ctx, conversationID, since, limit)
}

func (s *Service) InitRatchet(sharedSecret [32]byte, isInitiator bool, remotePub [32]byte) error {
	if isInitiator {
		dr, err := crypto.InitAlice(sharedSecret, remotePub)
		if err != nil { return err }
		*s.ratchet = *dr
	} else {
		dr, err := crypto.InitBob(sharedSecret, remotePub)
		if err != nil { return err }
		*s.ratchet = *dr
	}
	return nil
}

func (s *Service) EncryptMessage(plaintext []byte, ad []byte) ([]byte, error) {
	ciphertext, _, err := s.ratchet.Encrypt(plaintext, ad)
	return ciphertext, err
}

func (s *Service) DecryptMessage(ciphertext []byte, ad []byte) ([]byte, error) {
	plaintext, err := s.ratchet.Decrypt(ciphertext, [32]byte{}, ad)
	return plaintext, err
}

func signPayload(e EnvelopeData) []byte {
	b := make([]byte, 0, len(e.ConversationID)+len(e.MessageID)+8+len(e.Ciphertext))
	b = append(b, []byte(e.ConversationID)...)
	b = append(b, []byte(e.MessageID)...)
	var ts [8]byte
	binary.BigEndian.PutUint64(ts[:], uint64(e.SentAtUnix))
	b = append(b, ts[:]...)
	b = append(b, e.Ciphertext...)
	return b
}
