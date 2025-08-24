package p2p

import (
	"context"
	"log"
)

// Node - P2P узел
type Node struct {
	ctx        context.Context
	port       string
	nodeType   string
	nodeID     string
	logger     *log.Logger

	// TODO: Добавить LibP2P хост
	// TODO: Добавить DHT
	// TODO: Добавить протоколы
	// TODO: Добавить peer discovery
}

// NewNode - создаем новый P2P узел
func NewNode(ctx context.Context, port, nodeType string, logger *log.Logger) (*Node, error) {
	node := &Node{
		ctx:      ctx,
		port:     port,
		nodeType: nodeType,
		logger:   logger,
	}

	// TODO: Инициализировать LibP2P
	// TODO: Настроить протоколы
	// TODO: Подключиться к bootstrap узлам

	logger.Printf("P2P узел (%s) запущен на порту %s", nodeType, port)

	return node, nil
}

// Start - запускает P2P узел
func (n *Node) Start() error {
	n.logger.Println("Запускаем P2P узел")

	// TODO: Запустить LibP2P хост
	// TODO: Начать discovery
	// TODO: Подключиться к сети

	return nil
}

// Stop - останавливает P2P узел
func (n *Node) Stop() error {
	n.logger.Println("Останавливаем P2P узел")

	// TODO: Отключиться от сети
	// TODO: Закрыть соединения
	// TODO: Очистить ресурсы

	return nil
}

// SendMessage - отправляет сообщение через P2P
func (n *Node) SendMessage(toPeerID string, data []byte) error {
	n.logger.Printf("Отправляем сообщение пиру %s", toPeerID)

	// TODO: Найти пира в DHT
	// TODO: Установить соединение
	// TODO: Отправить данные
	// TODO: Обработать подтверждение

	return nil
}

// BroadcastMessage - рассылает сообщение всем подключенным пирам
func (n *Node) BroadcastMessage(data []byte) error {
	n.logger.Println("Рассылаем сообщение всем пирам")

	// TODO: Получить список активных пиров
	// TODO: Отправить каждому
	// TODO: Обработать ошибки

	return nil
}

// GetConnectedPeers - возвращает список подключенных пиров
func (n *Node) GetConnectedPeers() []string {
	// TODO: Получить список активных соединений

	return []string{}
}

// GetNodeID - возвращает ID узла
func (n *Node) GetNodeID() string {
	return n.nodeID
}
