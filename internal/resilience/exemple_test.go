package resilience_test

import (
	"context"
	"errors"
	"fmt"

	"github.com/runtime-metrics-course/internal/resilience"
)

func ExampleRetry() {
	// Simulating a database query function
	dbQuery := func() error {
		// In real code, this would be a call to pgx.Exec or similar
		return errors.New("connection reset by peer") // transient error
	}

	ctx := context.Background()
	err := resilience.Retry(ctx, dbQuery)
	fmt.Printf("Final error: %v\n", err)

	// Output:
	// Final error: operation failed after retries: connection reset by peer
}
