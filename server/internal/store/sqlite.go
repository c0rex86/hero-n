package store

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"dev.c0rex64.heroin/migrations"

	_ "modernc.org/sqlite"
)

type DB struct {
	SQL *sql.DB
}

func Open(ctx context.Context, dsn string) (*DB, error) {
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("open sqlite: %w", err)
	}
	if err := db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("ping sqlite: %w", err)
	}
	store := &DB{SQL: db}
	if err := store.migrate(ctx); err != nil {
		return nil, err
	}
	return store, nil
}

func (d *DB) migrate(ctx context.Context) error {
	entries, err := migrations.Files.ReadDir(".")
	if err != nil {
		return fmt.Errorf("read migrations: %w", err)
	}
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".sql") {
			continue
		}
		b, err := migrations.Files.ReadFile(e.Name())
		if err != nil {
			return fmt.Errorf("read migration %s: %w", e.Name(), err)
		}
		if _, err := d.SQL.ExecContext(ctx, string(b)); err != nil {
			return fmt.Errorf("apply migration %s: %w", e.Name(), err)
		}
	}
	return nil
}
