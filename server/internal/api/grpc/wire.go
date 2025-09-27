package grpcapi

import (
	"context"
	"time"

	"dev.c0rex64.heroin/internal/auth"
	"dev.c0rex64.heroin/internal/config"
	"dev.c0rex64.heroin/internal/store"
)

type Services struct {
	Auth *auth.Service
}

func BuildServices(ctx context.Context, cfg *config.Config, db *store.DB) (*Services, error) {
	repo := store.NewAuthRepo(db.SQL)
	h := auth.NewPasswordHasher(cfg.Security.KDF.Time, cfg.Security.KDF.MemoryMB, cfg.Security.KDF.Threads, cfg.Security.KDF.KeyLen)
	issuer := auth.NewTokenIssuer(cfg.PasetoKey(), cfg.Security.Token.Issuer, cfg.AccessTokenTTL())
	second := auth.NewSecondaryFactor(cfg.PasetoKey(), cfg.Security.SecondaryKey.Length, cfg.Security.SecondaryKey.RotateMinutes, cfg.Security.SecondaryKey.AllowedClockSkewSec)
	refreshTTL := time.Duration(cfg.Security.Token.RefreshDays) * 24 * time.Hour
	as := auth.NewService(repo, h, issuer, second, refreshTTL)
	return &Services{Auth: as}, nil
}
