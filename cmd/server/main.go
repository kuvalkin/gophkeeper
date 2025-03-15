package main

import (
	"context"
	"database/sql"
	"fmt"
	stdLog "log"
	"net"
	"os/signal"
	"syscall"

	"github.com/kuvalkin/gophkeeper/internal/server/service/user"
	userStorage "github.com/kuvalkin/gophkeeper/internal/server/storage/user"
	"github.com/kuvalkin/gophkeeper/internal/server/support/config"
	"github.com/kuvalkin/gophkeeper/internal/server/support/database"
	"github.com/kuvalkin/gophkeeper/internal/server/transport"
	"github.com/kuvalkin/gophkeeper/internal/support/log"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGQUIT, syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	err := log.InitLogger()
	if err != nil {
		stdLog.Fatal(fmt.Errorf("failed to initialize logger: %w", err))
	}

	defer func() {
		err = log.Logger().Sync()
		if err != nil {
			stdLog.Println(fmt.Errorf("failed to sync logger: %w", err))
		}
	}()

	conf, err := config.Resolve()
	if err != nil {
		log.Logger().Fatalw("failed to resolve config", "error", err)
	}

	db, err := initDB(ctx, *conf)
	if err != nil {
		log.Logger().Fatalw("failed to initialize database", "error", err)
	}

	services, err := initServices(ctx, *conf, db)
	if err != nil {
		log.Logger().Fatalw("failed to initialize services", "error", err)
	}

	// todo s3?
	serve(ctx, *conf, services)

	// if we are here, the server has been stopped
	log.Logger().Info("server shutdown complete")
}

func initDB(ctx context.Context, cnf config.Config) (*sql.DB, error) {
	log.Logger().Debug("connecting to DB")

	initCtx, cancel := context.WithTimeout(ctx, cnf.DatabaseTimeout)
	defer cancel()

	db, err := database.InitDB(initCtx, cnf.DatabaseDSN)
	if err != nil {
		return nil, fmt.Errorf("init db failed: %w", err)
	}

	migrateCtx, cancel := context.WithTimeout(context.Background(), cnf.DatabaseTimeout)
	defer cancel()

	err = database.Migrate(migrateCtx, db)
	if err != nil {
		return nil, fmt.Errorf("migrate failed: %w", err)
	}

	return db, nil
}

func initServices(_ context.Context, cnf config.Config, db *sql.DB) (transport.Services, error) {
	return transport.Services{
		User: user.NewService(
			userStorage.NewDatabaseRepository(db, cnf.DatabaseTimeout),
			user.Options{
				// todo
				TokenSecret:           []byte("todo"),
				PasswordSalt:          "todo",
				TokenExpirationPeriod: cnf.TokenExpirationPeriod,
			},
		),
	}, nil
}

func serve(ctx context.Context, conf config.Config, services transport.Services) {
	addr := conf.Address

	listener, err := net.Listen("tcp", addr)
	if err != nil {
		log.Logger().Fatalw("failed to listen on address", "address", addr, "error", err)
		return
	}

	srv, err := transport.NewServer(services)
	if err != nil {
		log.Logger().Fatalw("failed to create gRPC server", "error", err)
		return
	}

	go func() {
		log.Logger().Infow("starting gRPC server", "address", addr)

		if err = srv.Serve(listener); err != nil {
			log.Logger().Fatalw("error starting gRPC server", "error", err)
			return
		}
	}()

	// Waiting (indefinitely) for a signal
	<-ctx.Done()

	log.Logger().Info("shutting down server")

	srv.GracefulStop()
}
