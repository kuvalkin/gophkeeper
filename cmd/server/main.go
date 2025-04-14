package main

import (
	"context"
	"database/sql"
	"fmt"
	stdLog "log"
	"net"
	"os/signal"
	"syscall"

	"github.com/spf13/viper"
	"google.golang.org/grpc"

	"github.com/kuvalkin/gophkeeper/internal/server/service/entry"
	"github.com/kuvalkin/gophkeeper/internal/server/service/user"
	entryStorage "github.com/kuvalkin/gophkeeper/internal/server/storage/entry"
	userStorage "github.com/kuvalkin/gophkeeper/internal/server/storage/user"
	"github.com/kuvalkin/gophkeeper/internal/server/support/database"
	"github.com/kuvalkin/gophkeeper/internal/server/transport"
	"github.com/kuvalkin/gophkeeper/internal/storage/blob"
	"github.com/kuvalkin/gophkeeper/internal/support/log"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGQUIT, syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	err := log.InitServerLogger()
	if err != nil {
		stdLog.Fatal(fmt.Errorf("failed to initialize logger: %w", err))
	}

	defer func() {
		err = log.Logger().Sync()
		if err != nil {
			stdLog.Println(fmt.Errorf("failed to sync logger: %w", err))
		}
	}()

	config := newConfig()

	db, err := initDB(ctx, config)
	if err != nil {
		log.Logger().Fatalw("failed to initialize database", "error", err)
	}

	services, err := initServices(ctx, config, db)
	if err != nil {
		log.Logger().Fatalw("failed to initialize services", "error", err)
	}

	server, err := transport.NewServer(services, config.GetInt64("blob.chunk_size"))
	if err != nil {
		log.Logger().Fatalw("failed to initialize server", "error", err)
	}

	serve(ctx, config.GetString("address"), server)

	// if we are here, the server has been stopped
	log.Logger().Info("server shutdown complete")
}

func newConfig() *viper.Viper {
	config := viper.New()

	config.SetDefault("address", ":8080")
	config.MustBindEnv("address", "ADDRESS")

	config.MustBindEnv("token.secret", "TOKEN_SECRET")
	config.SetDefault("token.expiration", "720h")
	config.MustBindEnv("token.expiration", "TOKEN_EXPIRATION")

	config.MustBindEnv("password.salt", "PASSWORD_SALT")

	config.MustBindEnv("database.dsn", "DATABASE_DSN")

	config.MustBindEnv("blob.path", "BLOB_PATH")
	config.SetDefault("blob.chunk_size", 1024*1024) // 1MB
	config.MustBindEnv("blob.chunk_size", "BLOB_CHUNK_SIZE")

	return config
}

func initDB(ctx context.Context, config *viper.Viper) (*sql.DB, error) {
	log.Logger().Debug("connecting to DB")

	db, err := database.InitDB(ctx, config.GetString("database.dsn"))
	if err != nil {
		return nil, fmt.Errorf("init db failed: %w", err)
	}

	err = database.Migrate(ctx, db)
	if err != nil {
		return nil, fmt.Errorf("migrate failed: %w", err)
	}

	return db, nil
}

func initServices(_ context.Context, config *viper.Viper, db *sql.DB) (transport.Services, error) {
	br, err := blob.NewFileBlobRepository(config.GetString("blob.path"))
	if err != nil {
		return transport.Services{}, fmt.Errorf("failed to create blob repository: %w", err)
	}

	return transport.Services{
		User: user.NewService(
			userStorage.NewDatabaseRepository(db),
			user.Options{
				TokenSecret:           []byte(config.GetString("token.secret")),
				PasswordSalt:          config.GetString("password.salt"),
				TokenExpirationPeriod: config.GetDuration("token.expiration"),
			},
		),
		Entry: entry.New(
			entryStorage.NewDatabaseMetadataRepository(db),
			br,
		),
	}, nil
}

func serve(ctx context.Context, addr string, server *grpc.Server) {
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		log.Logger().Fatalw("failed to listen on address", "address", addr, "error", err)
		return
	}

	go func() {
		log.Logger().Infow("starting gRPC server", "address", addr)

		if err = server.Serve(listener); err != nil {
			log.Logger().Fatalw("error starting gRPC server", "error", err)
			return
		}
	}()

	// Waiting (indefinitely) for a signal
	<-ctx.Done()

	log.Logger().Info("shutting down server")

	server.GracefulStop()
}
