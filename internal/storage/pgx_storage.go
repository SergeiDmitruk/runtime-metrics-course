package storage

import (
	"database/sql"
)

type PgxStorage struct {
	*MemStorage
	conn *sql.DB
}

func NewPgxStorage(conn *sql.DB) *PgxStorage {
	return &PgxStorage{
		MemStorage: NewMemStorage(),
		conn:       conn,
	}
}

func (p *PgxStorage) Ping() error {
	return p.conn.Ping()
}
