package storage

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// PostgresConfig holds PostgreSQL connection configuration
type PostgresConfig struct {
	Host            string
	Port            int
	User            string
	Password        string
	Database        string
	SSLMode         string
	MaxConnections  int
	MinConnections  int
	ConnMaxLifetime time.Duration
	ConnMaxIdleTime time.Duration
}

// NewPostgresPool creates a new PostgreSQL connection pool
func NewPostgresPool(ctx context.Context, cfg PostgresConfig) (*pgxpool.Pool, error) {
	connString := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.Database, cfg.SSLMode,
	)

	poolConfig, err := pgxpool.ParseConfig(connString)
	if err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}

	// Set pool configuration
	if cfg.MaxConnections > 0 {
		poolConfig.MaxConns = int32(cfg.MaxConnections)
	} else {
		poolConfig.MaxConns = 10 // Default
	}

	if cfg.MinConnections > 0 {
		poolConfig.MinConns = int32(cfg.MinConnections)
	} else {
		poolConfig.MinConns = 2 // Default
	}

	if cfg.ConnMaxLifetime > 0 {
		poolConfig.MaxConnLifetime = cfg.ConnMaxLifetime
	} else {
		poolConfig.MaxConnLifetime = time.Hour // Default
	}

	if cfg.ConnMaxIdleTime > 0 {
		poolConfig.MaxConnIdleTime = cfg.ConnMaxIdleTime
	} else {
		poolConfig.MaxConnIdleTime = 30 * time.Minute // Default
	}

	// Create pool
	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		return nil, fmt.Errorf("create pool: %w", err)
	}

	// Test connection
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("ping database: %w", err)
	}

	return pool, nil
}

// MigrateUp runs database migrations
func MigrateUp(ctx context.Context, pool *pgxpool.Pool, migrationsPath string) error {
	// Read migration file
	// Note: In production, use a proper migration library like golang-migrate
	// For now, this is a simple implementation

	// Example: Read and execute migration files from migrationsPath
	// This should be implemented using a migration library

	return fmt.Errorf("migration not implemented - use golang-migrate or similar")
}

// MigrateDown rolls back database migrations
func MigrateDown(ctx context.Context, pool *pgxpool.Pool, migrationsPath string) error {
	return fmt.Errorf("migration not implemented - use golang-migrate or similar")
}
