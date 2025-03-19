package database

import (
	"context"
	"database/sql"
	"embed"
	"fmt"
	"net/url"
	"os"
	"path"

	"github.com/pressly/goose/v3"
	_ "modernc.org/sqlite"
)

//go:embed migrations/*.sql
var embedMigrations embed.FS

func InitDB(ctx context.Context, filePath string) (*sql.DB, error) {
	// todo may be move higher up?
	err := os.MkdirAll(path.Dir(filePath), 0600)
	if err != nil {
		return nil, fmt.Errorf("could not create database directory: %w", err)
	}

	u := &url.URL{
		Scheme:   "file",
		Path:     filePath,
		RawQuery: "_journal_mode=WAL&_auto_vacuum=1&_synchronous=FULL&mode=rwc&_busy_timeout=5000",
	}

	dsn := u.String()

	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("could not connect to database: %w", err)
	}

	err = db.PingContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("could not ping database: %w", err)
	}

	return db, nil
}

func Migrate(ctx context.Context, db *sql.DB) error {
	goose.SetBaseFS(embedMigrations)

	if err := goose.SetDialect("sqlite"); err != nil {
		return fmt.Errorf("could not set dialect: %w", err)
	}

	if err := goose.UpContext(ctx, db, "migrations"); err != nil {
		return fmt.Errorf("could not migrate database: %w", err)
	}

	return nil
}
