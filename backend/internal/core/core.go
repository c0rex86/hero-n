package core

import (
	"context"
	"crypto/rand"
	"log"

	"github.com/c0rex86/hero-n/backend/internal/crypto"
	"github.com/c0rex86/hero-n/backend/internal/protocol"
)

// Core - основное ядро приложения, координирует все компоненты
type Core struct {
	ctx       context.Context
	logger    *log.Logger
	crypto    *crypto.Engine  // шифрование
	protocol  *protocol.Engine // сетевой протокол
	nodeID    string          // уникальный ID узла
	nodeType  string          // тип узла (bootstrap/relay/storage)
}

// CoreConfig - настройки для ядра
type CoreConfig struct {
	NodeID   string
	NodeType string
}

// NewCore - создаем новое ядро приложения
func NewCore(ctx context.Context, logger *log.Logger) (*Core, error) {
	// Генерируем случайный ID для узла
	nodeID := make([]byte, 32)
	if _, err := rand.Read(nodeID); err != nil {
		return nil, err
	}

	// Запускаем криптографический движок
	cryptoEngine, err := crypto.NewEngine(logger)
	if err != nil {
		return nil, err
	}

	// Запускаем сетевой протокол
	protocolEngine, err := protocol.NewEngine(ctx, cryptoEngine, logger)
	if err != nil {
		return nil, err
	}

	core := &Core{
		ctx:      ctx,
		logger:   logger,
		crypto:   cryptoEngine,
		protocol: protocolEngine,
		nodeID:   string(nodeID),
		nodeType: "bootstrap", // по умолчанию
	}

	logger.Printf("Ядро запущено с ID узла: %s", core.nodeID)

	return core, nil
}

// GetNodeID - возвращает ID текущего узла
func (c *Core) GetNodeID() string {
	return c.nodeID
}

// GetNodeType - возвращает тип текущего узла
func (c *Core) GetNodeType() string {
	return c.nodeType
}

// ProcessMessage - обрабатывает входящие сообщения
func (c *Core) ProcessMessage(data []byte) ([]byte, error) {
	c.logger.Printf("Обрабатываем сообщение размером: %d байт", len(data))

	// Расшифровываем сообщение
	decrypted, err := c.crypto.Decrypt(data)
	if err != nil {
		c.logger.Printf("Не удалось расшифровать: %v", err)
		return nil, err
	}

	// Пропускаем через сетевой протокол
	result, err := c.protocol.ProcessMessage(decrypted)
	if err != nil {
		c.logger.Printf("Не удалось обработать: %v", err)
		return nil, err
	}

	// Шифруем ответ
	encrypted, err := c.crypto.Encrypt(result)
	if err != nil {
		c.logger.Printf("Не удалось зашифровать ответ: %v", err)
		return nil, err
	}

	return encrypted, nil
}

// GenerateKeyPair - генерирует новую пару ключей для узла
func (c *Core) GenerateKeyPair() error {
	return c.crypto.GenerateKeyPair()
}

// Close - останавливает ядро и очищает ресурсы
func (c *Core) Close() error {
	c.logger.Println("Останавливаем ядро...")

	if err := c.protocol.Close(); err != nil {
		c.logger.Printf("Ошибка остановки протокола: %v", err)
	}

	if err := c.crypto.Close(); err != nil {
		c.logger.Printf("Ошибка остановки шифрования: %v", err)
	}

	return nil
}
