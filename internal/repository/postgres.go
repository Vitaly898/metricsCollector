package repository

import (
	"context"
	"database/sql"
	"errors"
	"os"
	"path/filepath"
	"time"

	"github.com/golang-migrate/migrate/v4"
	pgxMigrate "github.com/golang-migrate/migrate/v4/database/pgx/v5"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"

	models "github.com/Vitaly898/metricsCollector/internal/model"
)

const dbTimeout = 15 * time.Second

var retryIntervals = []time.Duration{1 * time.Second, 3 * time.Second, 5 * time.Second}

type PostgresStorage struct {
	db *sql.DB
}

func isRetriableDBError(err error) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgerrcode.IsConnectionException(pgErr.Code)
	}
	return false
}

func retry(ctx context.Context, fn func() error) error {
	var err error
	for i := 0; i <= len(retryIntervals); i++ {
		err = fn()
		if err == nil || !isRetriableDBError(err) {
			return err
		}
		if i < len(retryIntervals) {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(retryIntervals[i]):
			}
		}
	}
	return err
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

	_ = retry(ctx, func() error {
		_, err := s.db.ExecContext(ctx, `
			INSERT INTO metrics (name, type, value)
			VALUES ($1, 'gauge', $2)
			ON CONFLICT (name, type) DO UPDATE SET value = EXCLUDED.value
		`, name, val)
		return err
	})
}

func (s *PostgresStorage) UpdateCounter(name string, val int64) {
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	_ = retry(ctx, func() error {
		_, err := s.db.ExecContext(ctx, `
			INSERT INTO metrics (name, type, delta)
			VALUES ($1, 'counter', $2)
			ON CONFLICT (name, type) DO UPDATE SET delta = COALESCE(metrics.delta, 0) + EXCLUDED.delta
		`, name, val)
		return err
	})
}

func (s *PostgresStorage) GetGauge(name string) (float64, bool) {
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	var value sql.NullFloat64
	err := retry(ctx, func() error {
		return s.db.QueryRowContext(ctx, `
			SELECT value FROM metrics WHERE name = $1 AND type = 'gauge'
		`, name).Scan(&value)
	})
	if err != nil || !value.Valid {
		return 0, false
	}
	return value.Float64, true
}

func (s *PostgresStorage) GetCounter(name string) (int64, bool) {
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	var delta sql.NullInt64
	err := retry(ctx, func() error {
		return s.db.QueryRowContext(ctx, `
			SELECT delta FROM metrics WHERE name = $1 AND type = 'counter'
		`, name).Scan(&delta)
	})
	if err != nil || !delta.Valid {
		return 0, false
	}
	return delta.Int64, true
}

func (s *PostgresStorage) GetAllGauge() map[string]float64 {
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	result := make(map[string]float64)
	var rows *sql.Rows
	err := retry(ctx, func() error {
		var err error
		rows, err = s.db.QueryContext(ctx, `
			SELECT name, value FROM metrics WHERE type = 'gauge'
		`)
		return err
	})
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
	var rows *sql.Rows
	err := retry(ctx, func() error {
		var err error
		rows, err = s.db.QueryContext(ctx, `
			SELECT name, delta FROM metrics WHERE type = 'counter'
		`)
		return err
	})
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

func (s *PostgresStorage) UpdateMetrics(metrics []models.Metrics) error {
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	return retry(ctx, func() error {
		tx, err := s.db.BeginTx(ctx, nil)
		if err != nil {
			return err
		}
		defer tx.Rollback()

		for _, m := range metrics {
			var delta *int64
			var value *float64
			switch m.MType {
			case models.Gauge:
				value = m.Value
			case models.Counter:
				delta = m.Delta
			}

			_, err := tx.ExecContext(ctx, `
				INSERT INTO metrics (name, type, delta, value)
				VALUES ($1, $2, $3, $4)
				ON CONFLICT (name, type) DO UPDATE SET
					delta = CASE WHEN EXCLUDED.type = 'counter' THEN COALESCE(metrics.delta, 0) + EXCLUDED.delta ELSE metrics.delta END,
					value = CASE WHEN EXCLUDED.type = 'gauge' THEN EXCLUDED.value ELSE metrics.value END
			`, m.ID, m.MType, delta, value)
			if err != nil {
				return err
			}
		}

		return tx.Commit()
	})
}
