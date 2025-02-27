package resilience

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/runtime-metrics-course/internal/logger"
)

func Retry(ctx context.Context, operation func() error) error {
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
			logger.Log.Sugar().Error("Retriable ошибка: %v. Повтор через %v...\n", err, delay)
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(delay):
			}
			continue
		}

		break
	}
	return fmt.Errorf("operation failed after retries: %w", err)
}
