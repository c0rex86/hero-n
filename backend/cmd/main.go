package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/c0rex86/hero-n/backend/internal/api"
	"github.com/c0rex86/hero-n/backend/internal/core"
	"github.com/c0rex86/hero-n/backend/internal/p2p"
	"github.com/c0rex86/hero-n/backend/internal/storage"
)

func main() {
	// Читаем параметры из командной строки
	var (
		port     = flag.String("port", "8080", "Порт для HTTP сервера")
		p2pPort  = flag.String("p2p-port", "4001", "Порт для P2P сети")
		nodeType = flag.String("type", "bootstrap", "Тип узла: bootstrap, relay, storage")
	)
	flag.Parse()

	// Настраиваем логирование
	logger := log.New(os.Stdout, "[HERO-N] ", log.LstdFlags|log.Lshortfile)

	// Создаем контекст для graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Инициализируем основные компоненты
	core, err := core.NewCore(ctx, logger)
	if err != nil {
		logger.Fatalf("Не удалось запустить ядро: %v", err)
	}

	// Запускаем P2P сеть
	p2pNode, err := p2p.NewNode(ctx, *p2pPort, *nodeType, logger)
	if err != nil {
		logger.Fatalf("Не удалось запустить P2P узел: %v", err)
	}

	// Настраиваем хранилище
	storage, err := storage.NewStorage(ctx, logger)
	if err != nil {
		logger.Fatalf("Не удалось запустить хранилище: %v", err)
	}

	// TODO: Создать сервисы для работы с данными
	// messageService := api.NewMessageService(core, core.GetCrypto(), storage, logger)
	// userService := api.NewUserService(storage, logger)
	// authService := api.NewAuthService(storage, logger)

	// Создаем HTTP API сервер
	server := api.NewServer(*port, core, p2pNode, storage, logger)

	// Запускаем сервисы
	logger.Printf("Запускаем HERO!N узел (тип: %s)", *nodeType)

	// P2P узел работает в фоне
	go func() {
		if err := p2pNode.Start(); err != nil {
			logger.Printf("Ошибка P2P узла: %v", err)
		}
	}()

	// HTTP сервер тоже в фоне
	go func() {
		logger.Printf("Запускаем HTTP сервер на порту %s", *port)
		if err := server.Start(); err != nil {
			logger.Printf("Ошибка HTTP сервера: %v", err)
		}
	}()

	// Ждем сигнала для graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Println("Выключаем HERO!N узел...")

	// Graceful shutdown - аккуратно останавливаем все
	cancel()

	// Останавливаем сервисы
	server.Stop()
	p2pNode.Stop()
	storage.Close()

	logger.Println("HERO!N узел остановлен")
}
''