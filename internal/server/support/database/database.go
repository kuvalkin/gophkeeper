package database

import (
	"context"
	"database/sql"
	"embed"
	"fmt"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
	"go.uber.org/zap"

	"github.com/kuvalkin/gophkeeper/internal/support/log"
)

//go:embed migrations/*.sql
var embedMigrations embed.FS

func InitDB(ctx context.Context, dsn string) (*sql.DB, error) {
	db, err := sql.Open("pgx", dsn)
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

	goose.SetLogger(&gooseLogger{
		log: log.Logger().Named("migrations"),
	})

	if err := goose.SetDialect("postgres"); err != nil {
		return fmt.Errorf("could not set dialect: %w", err)
	}

	if err := goose.UpContext(ctx, db, "migrations"); err != nil {
		return fmt.Errorf("could not migrate database: %w", err)
	}

	return nil
}

type gooseLogger struct {
	log *zap.SugaredLogger
}

func (g *gooseLogger) Fatalf(format string, v ...interface{}) {
	g.log.Fatalf(format, v...)
}

func (g *gooseLogger) Printf(format string, args ...interface{}) {
	g.log.Infof(format, args...)
}
