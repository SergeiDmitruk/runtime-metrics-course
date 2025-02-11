package utils

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/runtime-metrics-course/internal/models"
)

func MarshalMetricToJSON(mType, name string, val interface{}) (*models.MetricJSON, error) {
	metric := models.MetricJSON{
		ID:    name,
		MType: mType,
	}
	switch mType {
	case models.Counter:
		valInt, ok := val.(int64)
		if !ok {
			log.Println("parse error")
			return &metric, errors.New("parse error")
		}
		metric.Delta = &valInt
	case models.Gauge:

		valFl, ok := val.(float64)
		if !ok {
			log.Println("parse error")
			return &metric, errors.New("parse error")
		}
		metric.Value = &valFl
	default:
		log.Println("parse error")
		return &metric, errors.New("parse error")
	}
	return &metric, nil
}

func WithRetry(ctx context.Context, operation func() error) error {
	var err error
	var pgErr *pgconn.PgError
	var netErr net.Error
	delays := []time.Duration{1 * time.Second, 3 * time.Second, 5 * time.Second}

	for _, delay := range delays {
		err = operation()
		if err == nil {
			return nil
		}

		if errors.As(err, &pgErr) || errors.As(err, &netErr) || errors.Is(err, io.ErrUnexpectedEOF) {
			log.Printf("Retriable ошибка: %v. Повтор через %v...\n", err, delay)
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(delay):
			}
		}

		break
	}
	return fmt.Errorf("operation failed after retries: %w", err)
}
