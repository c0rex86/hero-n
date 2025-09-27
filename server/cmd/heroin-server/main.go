package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	grpcapi "dev.c0rex64.heroin/internal/api/grpc"
	httpapi "dev.c0rex64.heroin/internal/api/http"
	"dev.c0rex64.heroin/internal/config"
	"dev.c0rex64.heroin/internal/store"
)

func main() {
	cfgPath := os.Getenv("HEROIN_CONFIG")
	if cfgPath == "" { cfgPath = "configs/config.example.yaml" }
	cfg, err := config.Load(cfgPath)
	if err != nil { log.Fatalf("config: %v", err) }

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := config.SetupObservability(ctx, cfg.Observability); err != nil {
		log.Fatalf("observability: %v", err)
	}

	db, err := store.Open(ctx, cfg.Database.DSN)
	if err != nil { log.Fatalf("db: %v", err) }

	services, err := grpcapi.BuildServices(ctx, cfg, db)
	if err != nil { log.Fatalf("services: %v", err) }

	gs, err := grpcapi.New(cfg.Server.Listen.TCP)
	if err != nil { log.Fatalf("grpc: %v", err) }
	gs.AuthSvc = services.Auth
	gs.WireStorageAndMessaging(cfg.IPFS.Endpoint, cfg.IPFS.PinningEnabled, cfg.IPFS.ReplicationFactor, db.SQL)

	hs := httpapi.New(":8081")
	go func() { _ = hs.Start() }()
	go func() { _ = gs.Start() }()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh
	log.Println("shutdown")
}
