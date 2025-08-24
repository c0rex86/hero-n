package protocol

import (
	"context"
	"log"

	"github.com/c0re/dft/backend/internal/core"
	"github.com/c0re/dft/backend/internal/crypto"
)

// Engine - движок протокола сообщений
type Engine struct {
	ctx    context.Context
	core   *core.Core
	crypto *crypto.Engine
	logger *log.Logger

	// TODO: Добавить P2P соединения
	// TODO: Добавить обработчики сообщений
	// TODO: Добавить очередь сообщений
}

// NewEngine - создаем новый движок протокола
func NewEngine(ctx context.Context, core *core.Core, crypto *crypto.Engine, logger *log.Logger) (*Engine, error) {
	engine := &Engine{
		ctx:    ctx,
		core:   core,
		crypto: crypto,
		logger: logger,
	}

	logger.Println("Протокол запущен")

	return engine, nil
}

// ProcessMessage - обрабатывает входящее сообщение
func (e *Engine) ProcessMessage(data []byte) ([]byte, error) {
	e.logger.Printf("Обрабатываем сообщение размером %d байт", len(data))

	// TODO: Распарсить заголовок сообщения
	// TODO: Проверить подпись
	// TODO: Обработать в зависимости от типа

	// Пока просто возвращаем то же самое
	return data, nil
}

// SendMessage - отправляет сообщение другому узлу
func (e *Engine) SendMessage(toNodeID string, data []byte) error {
	e.logger.Printf("Отправляем сообщение узлу %s", toNodeID)

	// TODO: Найти соединение с узлом
	// TODO: Отправить через P2P
	// TODO: Обработать подтверждение доставки

	return nil
}

// BroadcastMessage - рассылает сообщение всем подключенным узлам
func (e *Engine) BroadcastMessage(data []byte) error {
	e.logger.Println("Рассылаем сообщение всем узлам")

	// TODO: Получить список активных соединений
	// TODO: Отправить всем узлам
	// TODO: Обработать ошибки

	return nil
}

// Close - останавливает протокол
func (e *Engine) Close() error {
	e.logger.Println("Останавливаем протокол")

	// TODO: Закрыть все соединения
	// TODO: Очистить ресурсы

	return nil
}
