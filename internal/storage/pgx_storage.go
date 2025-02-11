package storage

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/runtime-metrics-course/internal/models"
)

type PgxStorage struct {
	cache *MemStorage
	conn  *sql.DB
}

func NewPgxStorage(conn *sql.DB) *PgxStorage {
	s := &PgxStorage{
		cache: NewMemStorage(),
		conn:  conn,
	}
	s.InitCache(context.Background())
	return s
}

func (s *PgxStorage) Ping() error {
	return s.conn.Ping()
}

func (s *PgxStorage) UpdateGauge(ctx context.Context, name string, value float64) error {
	if _, err := s.conn.ExecContext(ctx,
		"INSERT INTO metrics (name, type, value, updated_at)"+
			"VALUES ($1, $2, $3, $4)"+
			" ON CONFLICT (name)"+" DO UPDATE SET value = $3, updated_at = $4 ", name, models.Gauge, value, time.Now().Format(time.RFC3339)); err != nil {
		return err
	}
	return s.cache.UpdateGauge(ctx, name, value)

}
func (s *PgxStorage) UpdateCounter(ctx context.Context, name string, delta int64) error {

	if _, err := s.conn.ExecContext(ctx,
		"INSERT INTO metrics (name, type, delta, updated_at)"+
			"VALUES ($1, $2, $3, $4)"+
			" ON CONFLICT (name)"+" DO UPDATE SET delta =  metrics.delta + $3, updated_at = $4 ", name, models.Counter, delta, time.Now().Format(time.RFC3339)); err != nil {
		return err
	}

	return s.cache.UpdateCounter(ctx, name, delta)
}

func (s *PgxStorage) GetMetrics(ctx context.Context) (models.Metrics, error) {
	return s.cache.GetMetrics(ctx)
}

func (s *PgxStorage) UpdateAll(ctx context.Context, metrics []models.MetricJSON) error {
	tx, err := s.conn.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()
	stmtCounter, err := tx.PrepareContext(ctx, "INSERT INTO metrics (name, type, delta, updated_at)"+
		"VALUES ($1, $2, $3, $4)"+
		" ON CONFLICT (name)"+" DO UPDATE SET delta =  metrics.delta + $3, updated_at = $4 ")
	if err != nil {
		return err
	}
	defer stmtCounter.Close()
	stmtGouge, err := tx.PrepareContext(ctx, "INSERT INTO metrics (name, type, value, updated_at)"+
		"VALUES ($1, $2, $3, $4)"+
		" ON CONFLICT (name)"+" DO UPDATE SET value = $3, updated_at = $4 ")
	if err != nil {
		return err
	}
	defer stmtGouge.Close()
	now := time.Now().Format(time.RFC3339)
	for _, metric := range metrics {
		switch {
		case metric.IsCounter() && metric.Delta != nil:
			if _, err := stmtCounter.ExecContext(ctx, metric.ID, metric.MType, metric.Delta, now); err != nil {
				return err
			}
		case metric.IsGauge() && metric.Value != nil:
			if _, err := stmtGouge.ExecContext(ctx, metric.ID, metric.MType, metric.Value, now); err != nil {
				return err
			}
		default:
			return fmt.Errorf("%s: invalid metric type or value", metric.ID)
		}

	}
	s.cache.UpdateAll(ctx, metrics)

	return tx.Commit()

}

func (s *PgxStorage) InitCache(ctx context.Context) error {
	rows, err := s.conn.QueryContext(ctx, "SELECT name, type, value, delta from  metrics")
	if err != nil {
		return err
	}
	defer rows.Close()
	var metrics []models.MetricJSON
	for rows.Next() {
		var m models.MetricJSON
		err = rows.Scan(&m.ID, &m.MType, &m.Value, &m.Delta)
		if err != nil {
			return err
		}

		metrics = append(metrics, m)
	}
	err = rows.Err()
	if err != nil {
		return err
	}

	for _, metric := range metrics {
		switch {
		case metric.IsCounter():
			s.cache.UpdateCounter(ctx, metric.ID, *metric.Delta)

		case metric.IsGauge():
			s.cache.UpdateGauge(ctx, metric.ID, *metric.Value)
		}
	}
	return nil
}
