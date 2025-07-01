// Package resilience provides utilities for handling transient failures in distributed systems.
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

// Retry executes an operation with exponential backoff retry logic for transient errors.
//
// The function will retry the operation up to 3 times with delays of 1s, 3s, and 5s between attempts
// if the error is determined to be transient. Transient errors include:
//   - PostgreSQL errors (*pgconn.PgError)
//   - Network errors (net.Error)
//   - Unexpected EOF errors (io.ErrUnexpectedEOF)
//
// Parameters:
//   - ctx: Context for cancellation and timeout control
//   - operation: The function to execute and retry on transient failures
//
// Returns:
//   - nil if the operation succeeds on any attempt
//   - The original error if a non-retriable error occurs
//   - ctx.Err() if the context is cancelled before completion
//   - A wrapped error with retry context if all attempts fail
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
