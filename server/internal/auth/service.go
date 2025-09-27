package auth

import (
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base32"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"golang.org/x/crypto/argon2"
)

type PasswordHasher struct {
	time     uint32
	memoryMB uint32
	threads  uint8
	keyLen   uint32
}

func NewPasswordHasher(time uint32, memoryMB uint32, threads uint8, keyLen uint32) PasswordHasher {
	return PasswordHasher{time: time, memoryMB: memoryMB, threads: threads, keyLen: keyLen}
}

func (h PasswordHasher) Hash(password []byte, salt []byte) []byte {
	return argon2.IDKey(password, salt, h.time, h.memoryMB*1024, h.threads, h.keyLen)
}

type TokenIssuer struct {
	key      []byte
	issuer   string
	lifetime time.Duration
}

type tokenPayload struct {
	Sub string `json:"sub"`
	Iss string `json:"iss"`
	Dev string `json:"device_id"`
	Exp int64  `json:"exp"`
	N   []byte `json:"n"`
}

func NewTokenIssuer(key []byte, issuer string, lifetime time.Duration) TokenIssuer {
	return TokenIssuer{key: key, issuer: issuer, lifetime: lifetime}
}

func (ti TokenIssuer) Issue(subject string, deviceID string, now time.Time) (string, time.Time, error) {
	exp := now.Add(ti.lifetime)
	p := tokenPayload{Sub: subject, Iss: ti.issuer, Dev: deviceID, Exp: exp.Unix(), N: randomBytes(8)}
	b, err := json.Marshal(p)
	if err != nil { return "", time.Time{}, err }
	sig := hmacSign(ti.key, b)
	tok := base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(append(b, sig...))
	return tok, exp, nil
}

func hmacSign(key, msg []byte) []byte {
	h := hmac.New(sha256.New, key)
	h.Write(msg)
	return h.Sum(nil)
}

type SecondaryFactor struct {
	secret []byte
	length int
	window time.Duration
	clockSkew time.Duration
}

func NewSecondaryFactor(secret []byte, length int, rotateMinutes int, clockSkewSec int) SecondaryFactor {
	return SecondaryFactor{secret: secret, length: length, window: time.Duration(rotateMinutes) * time.Minute, clockSkew: time.Duration(clockSkewSec) * time.Second}
}

func (sf SecondaryFactor) codeAt(t time.Time) string {
	counter := t.UTC().Unix() / int64(sf.window/time.Second)
	h := hmac.New(sha256.New, sf.secret)
	var b [8]byte
	for i := 0; i < 8; i++ {
		b[7-i] = byte(counter & 0xff)
		counter >>= 8
	}
	h.Write(b[:])
	d := h.Sum(nil)
	enc := strings.ToUpper(base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(d))
	if sf.length < len(enc) {
		enc = enc[:sf.length]
	}
	return enc
}

func (sf SecondaryFactor) Verify(now time.Time, code string) bool {
	code = strings.ToUpper(strings.TrimSpace(code))
	if code == "" { return false }
	if sf.length != len(code) { return false }
	for _, off := range []int64{-1, 0, 1} {
		cand := sf.codeAt(now.Add(time.Duration(off) * sf.window))
		if hmac.Equal([]byte(cand), []byte(code)) {
			return true
		}
	}
	return false
}

// Storage ports kept minimal for now

type UserRecord struct {
	ID                 string
	Username           string
	ServerSalt         []byte
	PasswordHash       []byte
	PublicKey          []byte
	SecondFactorSecret []byte
	CreatedAt          int64
}

type Store interface {
	CreateUser(ctx context.Context, u UserRecord) error
	FindUserByUsername(ctx context.Context, username string) (*UserRecord, error)
	CreateSession(ctx context.Context, userID string, deviceID string, refreshHash []byte, expiresAt time.Time) (string, error)
	GetSessionByRefreshHash(ctx context.Context, refreshHash []byte) (string, string, time.Time, error)
	WriteAudit(ctx context.Context, userID, deviceID, eventType string) error
	GetPublicKey(ctx context.Context, userID string) ([]byte, error)
}

var ErrInvalidCredentials = errors.New("invalid credentials")

// SRP/PAKE will be integrated later, we keep proof as opaque bytes

type Service struct {
	store      Store
	hasher     PasswordHasher
	issuer     TokenIssuer
	second     SecondaryFactor
	refreshTTL time.Duration
}

type AuditHook interface {
	OnLogin(userID, deviceID string)
	OnRefresh(userID, deviceID string)
}

func NewService(store Store, hasher PasswordHasher, issuer TokenIssuer, second SecondaryFactor, refreshTTL time.Duration) *Service {
	return &Service{store: store, hasher: hasher, issuer: issuer, second: second, refreshTTL: refreshTTL}
}

func (s *Service) Register(ctx context.Context, username string, passwordProof []byte, clientPub []byte) (string, error) {
	if username == "" { return "", fmt.Errorf("username empty") }
	parts := strings.Split(string(passwordProof), ":")
	if len(parts) != 2 { return "", fmt.Errorf("invalid password proof format") }
	salt, err := hex.DecodeString(parts[0])
	if err != nil { return "", err }
	hash, err := hex.DecodeString(parts[1])
	if err != nil { return "", err }
	id := generateUUIDv7()
	secret := randomStrongBytes(32)
	u := UserRecord{
		ID: id,
		Username: username,
		ServerSalt: salt,
		PasswordHash: hash,
		PublicKey: clientPub,
		SecondFactorSecret: secret,
		CreatedAt: time.Now().Unix(),
	}
	if err := s.store.CreateUser(ctx, u); err != nil { return "", err }
	return id, nil
}

func (s *Service) Login(ctx context.Context, username string, passwordProof []byte, deviceID string, secondCode string, now time.Time) (string, string, time.Time, error) {
	u, err := s.store.FindUserByUsername(ctx, username)
	if err != nil || u == nil { return "", "", time.Time{}, ErrInvalidCredentials }
	calc := s.hasher.Hash(passwordProof, u.ServerSalt)
	if !hmac.Equal(calc, u.PasswordHash) {
		return "", "", time.Time{}, ErrInvalidCredentials
	}
	if !s.second.Verify(now, secondCode) { return "", "", time.Time{}, ErrInvalidCredentials }
	access, exp, err := s.issuer.Issue(u.ID, deviceID, now)
	if err != nil { return "", "", time.Time{}, err }
	refreshStr, refreshHash := s.NewRefreshToken()
	refreshExp := now.Add(s.refreshTTL)
	if _, err := s.store.CreateSession(ctx, u.ID, deviceID, refreshHash, refreshExp); err != nil {
		return "", "", time.Time{}, err
	}
	return access, refreshStr, exp, nil
}

func (s *Service) NewRefreshToken() (string, []byte) {
	b := randomStrongBytes(32)
	tokenStr := base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(b)
	h := sha256.Sum256([]byte(tokenStr))
	return tokenStr, h[:]
}

func (s *Service) Refresh(ctx context.Context, refreshToken string, deviceID string, now time.Time) (string, time.Time, error) {
	h := sha256.Sum256([]byte(refreshToken))
	userID, storedDeviceID, sessExp, err := s.store.GetSessionByRefreshHash(ctx, h[:])
	if err != nil { return "", time.Time{}, ErrInvalidCredentials }
	if storedDeviceID != deviceID { return "", time.Time{}, ErrInvalidCredentials }
	if now.After(sessExp) { return "", time.Time{}, ErrInvalidCredentials }
	tok, exp, err := s.issuer.Issue(userID, deviceID, now)
	if err != nil { return "", time.Time{}, err }
	return tok, exp, nil
}

func (s *Service) GetPublicKey(ctx context.Context, userID string) ([]byte, error) {
	return s.store.GetPublicKey(ctx, userID)
}

// helpers

func randomBytes(n int) []byte {
	b := make([]byte, n)
	for i := range b { b[i] = byte(time.Now().UnixNano()>>uint(i%8)) }
	return b
}

func randomStrongBytes(n int) []byte {
	b := make([]byte, n)
	_, _ = rand.Read(b)
	return b
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
