package storage

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/runtime-metrics-course/internal/models"
)

// PgxStorage implements StorageIface using PostgreSQL as the backend storage
// with an in-memory cache for faster read operations.
type PgxStorage struct {
	cache *MemStorage // In-memory cache for quick access
	conn  *sql.DB     // PostgreSQL database connection
}

// NewPgxStorage creates a new PostgreSQL-backed storage with cache.
// Parameters:
//   - conn: Established database connection
//
// Returns:
//   - *PgxStorage: initialized storage instance with preloaded cache
func NewPgxStorage(conn *sql.DB) *PgxStorage {
	s := &PgxStorage{
		cache: NewMemStorage(),
		conn:  conn,
	}
	s.InitCache(context.Background())
	return s
}

// Ping checks the database connectivity.
// Implements StorageIface.Ping.
func (s *PgxStorage) Ping(ctx context.Context) error {
	return s.conn.PingContext(ctx)
}

// UpdateGauge stores or updates a gauge metric in both database and cache.
// Implements StorageIface.UpdateGauge.
func (s *PgxStorage) UpdateGauge(ctx context.Context, name string, value float64) error {
	_, err := s.conn.ExecContext(ctx,
		"INSERT INTO metrics (name, type, value, updated_at)"+
			"VALUES ($1, $2, $3, $4)"+
			" ON CONFLICT (name)"+
			" DO UPDATE SET value = $3, updated_at = $4",
		name, models.Gauge, value, time.Now().Format(time.RFC3339))
	if err != nil {
		return fmt.Errorf("failed to update gauge in database: %w", err)
	}

	return s.cache.UpdateGauge(ctx, name, value)
}

// UpdateCounter stores or increments a counter metric in both database and cache.
// Implements StorageIface.UpdateCounter.
func (s *PgxStorage) UpdateCounter(ctx context.Context, name string, delta int64) error {
	_, err := s.conn.ExecContext(ctx,
		"INSERT INTO metrics (name, type, delta, updated_at)"+
			"VALUES ($1, $2, $3, $4)"+
			" ON CONFLICT (name)"+
			" DO UPDATE SET delta = metrics.delta + $3, updated_at = $4",
		name, models.Counter, delta, time.Now().Format(time.RFC3339))
	if err != nil {
		return fmt.Errorf("failed to update counter in database: %w", err)
	}

	return s.cache.UpdateCounter(ctx, name, delta)
}

// GetMetrics retrieves all metrics from the cache.
// Implements StorageIface.GetMetrics.
func (s *PgxStorage) GetMetrics(ctx context.Context) (models.Metrics, error) {
	return s.cache.GetMetrics(ctx)
}

// UpdateAll performs atomic batch updates of multiple metrics.
// Implements StorageIface.UpdateAll.
func (s *PgxStorage) UpdateAll(ctx context.Context, metrics []models.MetricJSON) error {
	tx, err := s.conn.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	// Prepare statements for batch operations
	stmtCounter, err := tx.PrepareContext(ctx,
		"INSERT INTO metrics (name, type, delta, updated_at)"+
			"VALUES ($1, $2, $3, $4)"+
			" ON CONFLICT (name) DO UPDATE SET delta = metrics.delta + $3, updated_at = $4")
	if err != nil {
		return fmt.Errorf("failed to prepare counter statement: %w", err)
	}
	defer stmtCounter.Close()

	stmtGauge, err := tx.PrepareContext(ctx,
		"INSERT INTO metrics (name, type, value, updated_at)"+
			"VALUES ($1, $2, $3, $4)"+
			" ON CONFLICT (name) DO UPDATE SET value = $3, updated_at = $4")
	if err != nil {
		return fmt.Errorf("failed to prepare gauge statement: %w", err)
	}
	defer stmtGauge.Close()

	now := time.Now().Format(time.RFC3339)
	for _, metric := range metrics {
		switch {
		case metric.IsCounter() && metric.Delta != nil:
			if _, err := stmtCounter.ExecContext(ctx, metric.ID, metric.MType, *metric.Delta, now); err != nil {
				return fmt.Errorf("failed to update counter %s: %w", metric.ID, err)
			}
		case metric.IsGauge() && metric.Value != nil:
			if _, err := stmtGauge.ExecContext(ctx, metric.ID, metric.MType, *metric.Value, now); err != nil {
				return fmt.Errorf("failed to update gauge %s: %w", metric.ID, err)
			}
		default:
			return fmt.Errorf("%s: invalid metric type or value", metric.ID)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return s.cache.UpdateAll(ctx, metrics)
}

// InitCache loads all metrics from the database into memory cache.
// Called automatically during initialization.
func (s *PgxStorage) InitCache(ctx context.Context) error {
	rows, err := s.conn.QueryContext(ctx, "SELECT name, type, value, delta from metrics")
	if err != nil {
		return fmt.Errorf("failed to query metrics: %w", err)
	}
	defer rows.Close()

	var metrics []models.MetricJSON
	for rows.Next() {
		var m models.MetricJSON
		if err := rows.Scan(&m.ID, &m.MType, &m.Value, &m.Delta); err != nil {
			return fmt.Errorf("failed to scan metric row: %w", err)
		}
		metrics = append(metrics, m)
	}

	if err := rows.Err(); err != nil {
		return fmt.Errorf("row iteration error: %w", err)
	}

	// Update cache with loaded metrics
	for _, metric := range metrics {
		switch {
		case metric.IsCounter():
			if err := s.cache.UpdateCounter(ctx, metric.ID, *metric.Delta); err != nil {
				return fmt.Errorf("failed to update counter in cache: %w", err)
			}
		case metric.IsGauge():
			if err := s.cache.UpdateGauge(ctx, metric.ID, *metric.Value); err != nil {
				return fmt.Errorf("failed to update gauge in cache: %w", err)
			}
		}
	}
	return nil
}
