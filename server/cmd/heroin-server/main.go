package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	grpcapi "dev.c0rex64.heroin/internal/api/grpc"
	httpapi "dev.c0rex64.heroin/internal/api/http"
	"dev.c0rex64.heroin/internal/config"
	"dev.c0rex64.heroin/internal/discovery"
	"dev.c0rex64.heroin/internal/groups"
	"dev.c0rex64.heroin/internal/metrics"
	"dev.c0rex64.heroin/internal/p2p"
	"dev.c0rex64.heroin/internal/relay"
	"dev.c0rex64.heroin/internal/routing"
	"dev.c0rex64.heroin/internal/store"
	"dev.c0rex64.heroin/internal/transport"

	"github.com/libp2p/go-libp2p/core/peer"
)

func main() {
	// загружаем конфиг
	cfgPath := os.Getenv("HEROIN_CONFIG")
	if cfgPath == "" {
		cfgPath = "configs/config.example.yaml"
	}

	cfg, err := config.Load(cfgPath)
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// настраиваем observability
	if err := config.SetupObservability(ctx, cfg.Observability); err != nil {
		log.Fatalf("observability: %v", err)
	}

	// открываем бд
	db, err := store.Open(ctx, cfg.Database.DSN)
	if err != nil {
		log.Fatalf("db: %v", err)
	}
	defer db.Close()

	// создаем discovery сервис
	var disc *discovery.DHTDiscovery
	if cfg.Routing.EnableP2P {
		// парсим bootstrap пиры
		bootstrapPeers := make([]peer.AddrInfo, 0)
		for range cfg.Routing.BootstrapPeers {
			// тут должен быть парсинг multiaddr
			// пока пропускаем
		}
		
		discCfg := discovery.Config{
			ListenAddrs:    []string{"/ip4/0.0.0.0/tcp/4001"},
			BootstrapPeers: bootstrapPeers,
			Namespace:      "heroin",
			EnableMDNS:     true,
		}
		
		disc, err = discovery.NewDHTDiscovery(ctx, discCfg)
		if err != nil {
			log.Printf("p2p discovery failed: %v", err)
			// не критично, продолжаем без p2p
		} else {
			log.Printf("p2p discovery started: %s", disc.PeerID())
			defer disc.Close()
		}
	}

	// создаем менеджер транспортов
	tm := transport.NewManager()
	tm.AddTransport(transport.NewTCPTransport())
	tm.AddTransport(transport.NewQUICTransport(nil))
	tm.AddTransport(transport.NewWSSTransport())

	// создаем адаптивный роутер
	router := routing.NewAdaptiveRouter(
		cfg.Routing.MetricsWindowSec,
		cfg.Routing.LatencyThresholdMs,
	)

	// добавляем транспорты в роутер
	router.AddTransport(routing.Transport{
		ID:       "tcp",
		Type:     transport.TypeTCP,
		Endpoint: cfg.Server.Listen.TCP,
		Priority: 3,
	})
	
	if cfg.Routing.EnableQUIC {
		router.AddTransport(routing.Transport{
			ID:       "quic",
			Type:     transport.TypeQUIC,
			Endpoint: cfg.Server.Listen.QUIC,
			Priority: 1,
		})
	}

	// создаем метрики
	collector := metrics.NewCollector()

	// создаем p2p streaming
	p2pStream := p2p.NewStreamManager(disc.Host)

	// создаем relay chains
	relayMgr := relay.NewRelayManager()

	// инициализируем relay chains
	relayMgr.CreateChain(ctx, "main", 3)
	relayMgr.CreateChain(ctx, "backup", 2)

	// создаем сервис групп
	groupSvc := groups.NewService(db.SQL)

	// собираем сервисы
	services, err := grpcapi.BuildServices(ctx, cfg, db)
	if err != nil {
		log.Fatalf("services: %v", err)
	}

	// создаем grpc сервер
	gs, err := grpcapi.New(cfg.Server.Listen.TCP)
	if err != nil {
		log.Fatalf("grpc: %v", err)
	}

	// подключаем сервисы
	gs.AuthSvc = services.Auth
	gs.GroupSvc = groupSvc
	gs.StreamMgr = p2pStream
	gs.RelayMgr = relayMgr
	gs.Router = router
	gs.WireStorageAndMessaging(
		cfg.IPFS.Endpoint,
		cfg.IPFS.PinningEnabled,
		cfg.IPFS.ReplicationFactor,
		db.SQL,
		services,
		collector,
	)

	// http сервер
	hs := httpapi.New(":8081")

	// запускаем серверы
	go func() {
		log.Printf("starting http on :8081")
		if err := hs.Start(); err != nil {
			log.Printf("http server error: %v", err)
		}
	}()

	go func() {
		log.Printf("starting grpc on %s", cfg.Server.Listen.TCP)
		if err := gs.Start(); err != nil {
			log.Printf("grpc server error: %v", err)
		}
	}()

	// пробинг транспортов в фоне
	if router != nil {
		go func() {
			ticker := time.NewTicker(30 * time.Second)
			defer ticker.Stop()

			for {
				select {
				case <-ticker.C:
					router.ProbeTransport(ctx, "tcp")
					router.ProbeTransport(ctx, "quic")

					best := router.SelectBestTransport()
					if best != nil {
						tm.SetActive(best.Type)
					}
				case <-ctx.Done():
					return
				}
			}
		}()
	}

	// ротация relay chains
	relayMgr.StartRotation(ctx, 5*time.Minute)

	// периодический сбор метрик
	collector.StartPeriodicCollection(ctx, 1*time.Minute)

	// ждем сигнал завершения
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh

	log.Println("shutdown")
	cancel()

	// graceful shutdown
	gs.Stop()
	relayMgr.Close()
	if disc != nil {
		disc.Close()
	}
	p2pStream.Close()
}
