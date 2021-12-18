package store

import (
	"context"
	"time"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/log/zapadapter"
	"github.com/jackc/pgx/v4/pgxpool"
	"go.uber.org/zap"
)

type Store struct {
	db *pgxpool.Pool
}

func Connect(ctx context.Context, url string, logger *zap.Logger) (*Store, error) {
	cfg, err := pgxpool.ParseConfig(url)
	if err != nil {
		return nil, err
	}

	cfg.ConnConfig.LogLevel = pgx.LogLevelDebug
	cfg.ConnConfig.Logger = zapadapter.NewLogger(logger)

	cfg.MaxConns = 8
	cfg.MinConns = 4

	cfg.HealthCheckPeriod = 1 * time.Minute
	cfg.MaxConnLifetime = 24 * time.Hour
	cfg.MaxConnIdleTime = 30 * time.Minute
	cfg.ConnConfig.ConnectTimeout = 1 * time.Second

	dbpool, err := pgxpool.ConnectConfig(ctx, cfg)
	if err != nil {
		return nil, err
	}

	return &Store{
		db: dbpool,
	}, nil
}

func (s *Store) Connection() *pgxpool.Pool {
	return s.db
}

func (s *Store) Close() {
	s.db.Close()
}
