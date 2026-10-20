package db

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"regexp"

	"github.com/jmoiron/sqlx"
	"github.com/pressly/goose/v3"
	"github.com/pressly/goose/v3/database"
)

var schemaRe = regexp.MustCompile(`^[a-z_]+$`)

// Migrate создаёт схему и применяет SQL-миграции из fsys (каталог "migrations").
// У каждой схемы своя таблица версий <schema>.goose_db_version, поэтому зоны
// мигрируются независимо. Отсутствие миграций — не ошибка.
func Migrate(ctx context.Context, d *sqlx.DB, schema string, fsys fs.FS) error {
	if !schemaRe.MatchString(schema) {
		return fmt.Errorf("недопустимое имя схемы %q", schema)
	}
	if _, err := d.ExecContext(ctx, `CREATE SCHEMA IF NOT EXISTS `+schema); err != nil {
		return fmt.Errorf("create schema %s: %w", schema, err)
	}
	store, err := database.NewStore(goose.DialectPostgres, schema+".goose_db_version")
	if err != nil {
		return err
	}
	sub, err := fs.Sub(fsys, "migrations")
	if err != nil {
		return err
	}
	p, err := goose.NewProvider("", d.DB, sub, goose.WithStore(store))
	if err != nil {
		if errors.Is(err, goose.ErrNoMigrations) {
			return nil
		}
		return fmt.Errorf("migrate %s: %w", schema, err)
	}
	if _, err := p.Up(ctx); err != nil && !errors.Is(err, goose.ErrNoMigrations) {
		return fmt.Errorf("migrate %s: %w", schema, err)
	}
	return nil
}
