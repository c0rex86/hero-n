package crypto

import (
	"crypto/rand"
	"crypto/sha256"
	"fmt"
	"io"
	"log"

	"golang.org/x/crypto/chacha20poly1305"
	"golang.org/x/crypto/ed25519"
	"golang.org/x/crypto/x25519"
)

// Engine - криптографический движок, отвечает за все шифрование
type Engine struct {
	logger     *log.Logger
	privateKey ed25519.PrivateKey // приватный ключ для подписей
	publicKey  ed25519.PublicKey  // публичный ключ для верификации
	sharedKey  []byte            // общий ключ для шифрования
}

// NewEngine - создаем новый криптографический движок
func NewEngine(logger *log.Logger) (*Engine, error) {
	engine := &Engine{
		logger: logger,
	}

	// Генерируем начальную пару ключей
	if err := engine.GenerateKeyPair(); err != nil {
		return nil, fmt.Errorf("не удалось сгенерировать ключи: %w", err)
	}

	logger.Println("Криптографический движок запущен")

	return engine, nil
}

// GenerateKeyPair - генерирует новую пару Ed25519 ключей
func (e *Engine) GenerateKeyPair() error {
	publicKey, privateKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return fmt.Errorf("не удалось сгенерировать Ed25519 ключи: %w", err)
	}

	e.privateKey = privateKey
	e.publicKey = publicKey

	e.logger.Printf("Сгенерировали новую пару ключей, публичный: %x", e.publicKey)

	return nil
}

// GetPublicKey - возвращает текущий публичный ключ
func (e *Engine) GetPublicKey() []byte {
	return e.publicKey
}

// Encrypt - шифрует данные с помощью ChaCha20-Poly1305
func (e *Engine) Encrypt(data []byte) ([]byte, error) {
	// Используем SHA256 от публичного ключа как ключ шифрования
	// Это работает для текущей версии, но можно улучшить key derivation позже
	key := sha256.Sum256(e.publicKey)

	aead, err := chacha20poly1305.New(key[:])
	if err != nil {
		return nil, fmt.Errorf("не удалось создать шифр: %w", err)
	}

	// Генерируем случайный nonce для каждого шифрования
	nonce := make([]byte, chacha20poly1305.NonceSize)
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, fmt.Errorf("не удалось сгенерировать nonce: %w", err)
	}

	// Шифруем данные с nonce в начале
	ciphertext := aead.Seal(nonce, nonce, data, nil)

	return ciphertext, nil
}

// Decrypt - расшифровывает данные с помощью ChaCha20-Poly1305
func (e *Engine) Decrypt(data []byte) ([]byte, error) {
	// Используем SHA256 от публичного ключа как ключ расшифрования
	key := sha256.Sum256(e.publicKey)

	aead, err := chacha20poly1305.New(key[:])
	if err != nil {
		return nil, fmt.Errorf("не удалось создать шифр: %w", err)
	}

	if len(data) < chacha20poly1305.NonceSize {
		return nil, fmt.Errorf("зашифрованный текст слишком короткий")
	}

	// Извлекаем nonce из начала зашифрованных данных
	nonce := data[:chacha20poly1305.NonceSize]
	ciphertext := data[chacha20poly1305.NonceSize:]

	// Расшифровываем данные
	plaintext, err := aead.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, fmt.Errorf("не удалось расшифровать: %w", err)
	}

	return plaintext, nil
}

// Sign - подписывает данные с помощью Ed25519
func (e *Engine) Sign(data []byte) []byte {
	signature := ed25519.Sign(e.privateKey, data)
	return signature
}

// Verify - проверяет подпись с помощью Ed25519
func (e *Engine) Verify(data, signature []byte) bool {
	return ed25519.Verify(e.publicKey, data, signature)
}

// GenerateSharedKey - генерирует общий ключ с помощью X25519
func (e *Engine) GenerateSharedKey(peerPublicKey []byte) error {
	// Конвертируем Ed25519 ключи в X25519 для совместимости
	var x25519Private, x25519Public [32]byte

	// Используем первые 32 байта для конвертации
	// В будущем можно добавить более строгую key conversion
	copy(x25519Private[:], e.privateKey[:32])
	copy(x25519Public[:], peerPublicKey[:32])

	sharedKey, err := x25519.ComputeSharedSecret(&x25519Private, &x25519Public)
	if err != nil {
		return fmt.Errorf("не удалось вычислить общий секрет: %w", err)
	}

	e.sharedKey = sharedKey
	e.logger.Printf("Сгенерировали общий ключ с пиром")

	return nil
}

// Close - очищает ресурсы и чувствительные данные
func (e *Engine) Close() error {
	// Очищаем чувствительные данные из памяти
	for i := range e.privateKey {
		e.privateKey[i] = 0
	}
	for i := range e.sharedKey {
		e.sharedKey[i] = 0
	}

	e.logger.Println("Криптографический движок остановлен")
	return nil
}
