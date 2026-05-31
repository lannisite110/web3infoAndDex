package db

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

// MySQL wraps a sql.DB connection pool for Railway or local MySQL.
type MySQL struct {
	DB *sql.DB
}

// ConnectMySQL opens a MySQL pool and verifies with ping.
func ConnectMySQL(ctx context.Context, dsn string) (*MySQL, error) {
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, fmt.Errorf("mysql open: %w", err)
	}
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	pingCtx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()
	if err := db.PingContext(pingCtx); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("mysql ping: %w", err)
	}
	return &MySQL{DB: db}, nil
}

// Ping checks database connectivity.
func (m *MySQL) Ping(ctx context.Context) error {
	return m.DB.PingContext(ctx)
}

// Close closes the connection pool.
func (m *MySQL) Close() error {
	return m.DB.Close()
}
