package storage

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
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

func TestUpdateGaugeDB_Error(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	mockStorage := &PgxStorage{conn: db}

	name := "cpu_usage"
	value := 42.5
	now := time.Now().Format(time.RFC3339)

	mock.ExpectExec("INSERT INTO metrics").
		WithArgs(name, "gauge", value, now).
		WillReturnError(sql.ErrConnDone)

	err = mockStorage.UpdateGauge(context.Background(), name, value)
	assert.Error(t, err)
	assert.Equal(t, sql.ErrConnDone, err)
}

func TestUpdateCounter_Error(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	mockStorage := &PgxStorage{conn: db}

	name := "requests_total"
	delta := int64(5)
	now := time.Now().Format(time.RFC3339)

	mock.ExpectExec("INSERT INTO metrics").
		WithArgs(name, "counter", delta, now).
		WillReturnError(sql.ErrConnDone)

	err = mockStorage.UpdateCounter(context.Background(), name, delta)
	assert.Error(t, err)
	assert.Equal(t, sql.ErrConnDone, err)
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

	rows := sqlmock.NewRows([]string{"id", "type", "value", "delta"}).
		AddRow("metric1", "gauge", 123.45, nil).
		AddRow("metric2", "counter", nil, 10)

	mock.ExpectQuery("SELECT id, type, value, delta from metrics").
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
