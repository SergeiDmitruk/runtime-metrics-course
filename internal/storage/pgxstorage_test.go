package storage

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/runtime-metrics-course/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUpdateGaugeDB(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	mockStorage := &PgxStorage{
		conn:  db,
		cache: NewMemStorage(),
	}

	name := "cpu_usage"
	value := 42.5
	now := time.Now().Format(time.RFC3339)

	mock.ExpectExec("INSERT INTO metrics").
		WithArgs(name, "gauge", value, now).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err = mockStorage.UpdateGauge(context.Background(), name, value)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUpdateCounterDB(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	mockStorage := &PgxStorage{
		conn:  db,
		cache: NewMemStorage(),
	}

	name := "requests_total"
	delta := int64(5)
	now := time.Now().Format(time.RFC3339)

	mock.ExpectExec("INSERT INTO metrics").
		WithArgs(name, "counter", delta, now).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err = mockStorage.UpdateCounter(context.Background(), name, delta)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUpdateGauge_Error(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	mockStorage := &PgxStorage{conn: db, cache: NewMemStorage()}

	name := "cpu_usage"
	value := 42.5
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	mock.ExpectExec("INSERT INTO metrics").
		WithArgs(name, models.Gauge, value, sqlmock.AnyArg()).
		WillReturnError(sql.ErrConnDone)

	err = mockStorage.UpdateGauge(ctx, name, value)

	expectedErr := "sql: connection is already closed"
	assert.EqualError(t, err, expectedErr)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUpdateCounter_Error(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	mockStorage := &PgxStorage{conn: db, cache: NewMemStorage()}

	name := "requests_total"
	delta := int64(5)

	mock.ExpectExec("INSERT INTO metrics").
		WithArgs(name, models.Counter, delta, sqlmock.AnyArg()).
		WillReturnError(errors.New("failed to execute query"))

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	err = mockStorage.UpdateCounter(ctx, name, delta)

	expectedErr := "failed to execute query"
	assert.EqualError(t, err, expectedErr)

	assert.NoError(t, mock.ExpectationsWereMet())
}
func TestPgxStorage_InitCache(t *testing.T) {
	mockDB, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer mockDB.Close()

	s := &PgxStorage{
		conn:  mockDB,
		cache: NewMemStorage(),
	}

	ctx := context.Background()

	rows := sqlmock.NewRows([]string{"name", "type", "value", "delta"}).
		AddRow("metric1", "gauge", 123.45, nil).
		AddRow("metric2", "counter", nil, 10)

	mock.ExpectQuery("SELECT name, type, value, delta from metrics").
		WillReturnRows(rows)

	err = s.InitCache(ctx)
	assert.NoError(t, err)

	gaugeValue, exists := s.cache.gauges["metric1"]
	assert.True(t, exists)
	assert.Equal(t, 123.45, gaugeValue)

	counterValue, exists := s.cache.counters["metric2"]
	assert.True(t, exists)
	assert.Equal(t, int64(10), counterValue)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestPgxStorage_UpdateAll(t *testing.T) {
	t.Run("successful update", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer db.Close()

		pgStorage := NewPgxStorage(db)
		ctx := context.Background()
		now := time.Now().Format(time.RFC3339)

		metrics := []models.MetricJSON{
			{ID: "requests", MType: "counter", Delta: int64Ptr(42)},
			{ID: "temperature", MType: "gauge", Value: float64Ptr(25.5)},
		}

		mock.ExpectBegin()
		mock.ExpectPrepare("INSERT INTO metrics .* DO UPDATE SET delta")
		mock.ExpectPrepare("INSERT INTO metrics .* DO UPDATE SET value")

		mock.ExpectExec("INSERT INTO metrics .* DO UPDATE SET delta").
			WithArgs("requests", "counter", *metrics[0].Delta, now).
			WillReturnResult(sqlmock.NewResult(1, 1))

		mock.ExpectExec("INSERT INTO metrics .* DO UPDATE SET value").
			WithArgs("temperature", "gauge", *metrics[1].Value, now).
			WillReturnResult(sqlmock.NewResult(1, 1))

		mock.ExpectCommit()

		err = pgStorage.UpdateAll(ctx, metrics)
		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("transaction begin error", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer db.Close()

		pgStorage := NewPgxStorage(db)
		ctx := context.Background()

		mock.ExpectBegin().WillReturnError(errors.New("begin error"))

		err = pgStorage.UpdateAll(ctx, nil)
		assert.ErrorContains(t, err, "begin error")
	})

	t.Run("prepare statement error", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer db.Close()

		pgStorage := NewPgxStorage(db)
		ctx := context.Background()

		mock.ExpectBegin()
		mock.ExpectPrepare("INSERT INTO metrics .* DO UPDATE SET delta").
			WillReturnError(errors.New("prepare error"))

		err = pgStorage.UpdateAll(ctx, nil)
		assert.ErrorContains(t, err, "prepare error")
	})

	t.Run("exec error", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer db.Close()

		pgStorage := NewPgxStorage(db)
		ctx := context.Background()
		now := time.Now().Format(time.RFC3339)

		metrics := []models.MetricJSON{
			{ID: "requests", MType: "counter", Delta: int64Ptr(42)},
		}

		mock.ExpectBegin()
		mock.ExpectPrepare("INSERT INTO metrics .* DO UPDATE SET delta")
		mock.ExpectPrepare("INSERT INTO metrics .* DO UPDATE SET value")

		mock.ExpectExec("INSERT INTO metrics .* DO UPDATE SET delta").
			WithArgs("requests", "counter", *metrics[0].Delta, now).
			WillReturnError(errors.New("exec error"))

		mock.ExpectRollback()

		err = pgStorage.UpdateAll(ctx, metrics)
		assert.ErrorContains(t, err, "exec error")
	})

	t.Run("commit error", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer db.Close()

		pgStorage := NewPgxStorage(db)
		ctx := context.Background()
		now := time.Now().Format(time.RFC3339)

		metrics := []models.MetricJSON{
			{ID: "requests", MType: "counter", Delta: int64Ptr(42)},
		}

		mock.ExpectBegin()
		mock.ExpectPrepare("INSERT INTO metrics .* DO UPDATE SET delta")
		mock.ExpectPrepare("INSERT INTO metrics .* DO UPDATE SET value")

		mock.ExpectExec("INSERT INTO metrics .* DO UPDATE SET delta").
			WithArgs("requests", "counter", *metrics[0].Delta, now).
			WillReturnResult(sqlmock.NewResult(1, 1))

		mock.ExpectCommit().WillReturnError(errors.New("commit error"))

		err = pgStorage.UpdateAll(ctx, metrics)
		assert.ErrorContains(t, err, "commit error")
	})

}

func int64Ptr(i int64) *int64       { return &i }
func float64Ptr(f float64) *float64 { return &f }

func BenchmarkPgxStorage_UpdateGauge(b *testing.B) {
	db, mock, err := sqlmock.New()
	if err != nil {
		b.Fatal(err)
	}
	defer db.Close()

	storage := NewPgxStorage(db)
	ctx := context.Background()
	name := "test_gauge"
	value := 42.5

	// Настраиваем мок для каждого вызова
	for i := 0; i < b.N; i++ {
		mock.ExpectExec("INSERT INTO metrics").
			WithArgs(name, models.Gauge, value, sqlmock.AnyArg()).
			WillReturnResult(sqlmock.NewResult(1, 1))
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if err := storage.UpdateGauge(ctx, name, value); err != nil {
			b.Error(err)
		}
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		b.Error(err)
	}
}

func BenchmarkPgxStorage_UpdateCounter(b *testing.B) {
	db, mock, err := sqlmock.New()
	if err != nil {
		b.Fatal(err)
	}
	defer db.Close()

	storage := NewPgxStorage(db)
	ctx := context.Background()
	name := "test_counter"
	delta := int64(1)

	// Настраиваем мок для каждого вызова
	for i := 0; i < b.N; i++ {
		mock.ExpectExec("INSERT INTO metrics").
			WithArgs(name, models.Counter, delta, sqlmock.AnyArg()).
			WillReturnResult(sqlmock.NewResult(1, 1))
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if err := storage.UpdateCounter(ctx, name, delta); err != nil {
			b.Error(err)
		}
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		b.Error(err)
	}
}

func BenchmarkPgxStorage_UpdateAll(b *testing.B) {
	db, mock, err := sqlmock.New()
	if err != nil {
		b.Fatal(err)
	}
	defer db.Close()

	storage := NewPgxStorage(db)
	ctx := context.Background()

	metrics := []models.MetricJSON{
		{ID: "counter1", MType: "counter", Delta: int64Ptr(10)},
		{ID: "gauge1", MType: "gauge", Value: float64Ptr(1.23)},
	}

	// Настраиваем моки для всех итераций
	for i := 0; i < b.N; i++ {
		mock.ExpectBegin()
		mock.ExpectPrepare("INSERT INTO metrics .* DO UPDATE SET delta")
		mock.ExpectPrepare("INSERT INTO metrics .* DO UPDATE SET value")
		mock.ExpectExec("INSERT INTO metrics .* DO UPDATE SET delta").
			WithArgs("counter1", "counter", int64(10), sqlmock.AnyArg()).
			WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectExec("INSERT INTO metrics .* DO UPDATE SET value").
			WithArgs("gauge1", "gauge", 1.23, sqlmock.AnyArg()).
			WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if err := storage.UpdateAll(ctx, metrics); err != nil {
			b.Error(err)
		}
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		b.Error(err)
	}
}
