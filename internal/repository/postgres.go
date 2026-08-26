package repository

import (
	"context"
	"database/sql"
	"os"
	"path/filepath"
	"time"

	"github.com/golang-migrate/migrate/v4"
	pgxMigrate "github.com/golang-migrate/migrate/v4/database/pgx/v5"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

const dbTimeout = 3 * time.Second

type PostgresStorage struct {
	db *sql.DB
}

func NewPostgresStorage(db *sql.DB) (*PostgresStorage, error) {
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		return nil, err
	}

	driver, err := pgxMigrate.WithInstance(db, &pgxMigrate.Config{})
	if err != nil {
		return nil, err
	}

	wd, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	m, err := migrate.NewWithDatabaseInstance(
		"file://"+filepath.Join(wd, "migrations"),
		"pgx",
		driver,
	)
	if err != nil {
		return nil, err
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return nil, err
	}

	return &PostgresStorage{db: db}, nil
}

func (s *PostgresStorage) UpdateGauge(name string, val float64) {
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	_, _ = s.db.ExecContext(ctx, `
		INSERT INTO metrics (name, type, value)
		VALUES ($1, 'gauge', $2)
		ON CONFLICT (name, type) DO UPDATE SET value = EXCLUDED.value
	`, name, val)
}

func (s *PostgresStorage) UpdateCounter(name string, val int64) {
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	_, _ = s.db.ExecContext(ctx, `
		INSERT INTO metrics (name, type, delta)
		VALUES ($1, 'counter', $2)
		ON CONFLICT (name, type) DO UPDATE SET delta = COALESCE(metrics.delta, 0) + EXCLUDED.delta
	`, name, val)
}

func (s *PostgresStorage) GetGauge(name string) (float64, bool) {
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	var value sql.NullFloat64
	err := s.db.QueryRowContext(ctx, `
		SELECT value FROM metrics WHERE name = $1 AND type = 'gauge'
	`, name).Scan(&value)
	if err != nil || !value.Valid {
		return 0, false
	}
	return value.Float64, true
}

func (s *PostgresStorage) GetCounter(name string) (int64, bool) {
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	var delta sql.NullInt64
	err := s.db.QueryRowContext(ctx, `
		SELECT delta FROM metrics WHERE name = $1 AND type = 'counter'
	`, name).Scan(&delta)
	if err != nil || !delta.Valid {
		return 0, false
	}
	return delta.Int64, true
}

func (s *PostgresStorage) GetAllGauge() map[string]float64 {
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	result := make(map[string]float64)
	rows, err := s.db.QueryContext(ctx, `
		SELECT name, value FROM metrics WHERE type = 'gauge'
	`)
	if err != nil {
		return result
	}
	defer rows.Close()

	for rows.Next() {
		var name string
		var value float64
		if err := rows.Scan(&name, &value); err == nil {
			result[name] = value
		}
	}

	return result
}

func (s *PostgresStorage) GetAllCounter() map[string]int64 {
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	result := make(map[string]int64)
	rows, err := s.db.QueryContext(ctx, `
		SELECT name, delta FROM metrics WHERE type = 'counter'
	`)
	if err != nil {
		return result
	}
	defer rows.Close()

	for rows.Next() {
		var name string
		var delta int64
		if err := rows.Scan(&name, &delta); err == nil {
			result[name] = delta
		}
	}

	return result
}
