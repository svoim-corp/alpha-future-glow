// Package db — подключение к PostgreSQL, транзакции и миграции по схемам.
package db

import (
	"context"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"

	_ "github.com/jackc/pgx/v5/stdlib" // драйвер "pgx"
)

// Open подключается к БД, повторяя пинг (БД может стартовать дольше сервера).
func Open(ctx context.Context, url string) (*sqlx.DB, error) {
	d, err := sqlx.Open("pgx", url)
	if err != nil {
		return nil, err
	}
	d.SetMaxOpenConns(20)
	d.SetMaxIdleConns(5)
	d.SetConnMaxLifetime(30 * time.Minute)

	var lastErr error
	for i := 0; i < 10; i++ {
		pctx, cancel := context.WithTimeout(ctx, 3*time.Second)
		lastErr = d.PingContext(pctx)
		cancel()
		if lastErr == nil {
			return d, nil
		}
		select {
		case <-ctx.Done():
			_ = d.Close()
			return nil, ctx.Err()
		case <-time.After(time.Second):
		}
	}
	_ = d.Close()
	return nil, fmt.Errorf("db ping: %w", lastErr)
}

// WithTx выполняет fn в транзакции: commit при успехе, rollback при ошибке или панике.
func WithTx(ctx context.Context, d *sqlx.DB, fn func(*sqlx.Tx) error) (err error) {
	tx, err := d.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback()
			panic(p)
		}
		if err != nil {
			_ = tx.Rollback()
		}
	}()
	if err = fn(tx); err != nil {
		return err
	}
	return tx.Commit()
}
