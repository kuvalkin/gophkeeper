package main

import (
	"context"
	"database/sql"
	"fmt"
	stdLog "log"
	"net"
	"os/signal"
	"syscall"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/spf13/viper"

	"github.com/kuvalkin/gophkeeper/internal/server/service/sync"
	"github.com/kuvalkin/gophkeeper/internal/server/service/user"
	sync2 "github.com/kuvalkin/gophkeeper/internal/server/storage/sync"
	userStorage "github.com/kuvalkin/gophkeeper/internal/server/storage/user"
	"github.com/kuvalkin/gophkeeper/internal/server/support/database"
	"github.com/kuvalkin/gophkeeper/internal/server/transport"
	"github.com/kuvalkin/gophkeeper/internal/support/log"
	"github.com/kuvalkin/gophkeeper/internal/support/transaction"
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

	config := viper.New()
	config.AutomaticEnv()
	config.SetDefault("ADDRESS", "localhost:8080")
	config.SetDefault("TOKEN_EXPIRATION_PERIOD", "30d")

	db, err := initDB(ctx, config)
	if err != nil {
		log.Logger().Fatalw("failed to initialize database", "error", err)
	}

	s3, err := minio.New(
		config.GetString("S3_ENDPOINT"),
		&minio.Options{
			Creds:  credentials.NewStaticV4(config.GetString("S3_ACCESS_KEY"), config.GetString("S3_SECRET_KEY"), ""),
			Secure: config.GetBool("S3_SECURE"),
		},
	)
	if err != nil {
		log.Logger().Fatalw("failed to initialize S3 client", "error", err)
	}

	services, err := initServices(ctx, config, db, s3)
	if err != nil {
		log.Logger().Fatalw("failed to initialize services", "error", err)
	}

	serve(ctx, config.GetString("ADDRESS"), services)

	// if we are here, the server has been stopped
	log.Logger().Info("server shutdown complete")
}

func initDB(ctx context.Context, cnf *viper.Viper) (*sql.DB, error) {
	log.Logger().Debug("connecting to DB")

	db, err := database.InitDB(ctx, cnf.GetString("DATABASE_DSN"))
	if err != nil {
		return nil, fmt.Errorf("init db failed: %w", err)
	}

	err = database.Migrate(ctx, db)
	if err != nil {
		return nil, fmt.Errorf("migrate failed: %w", err)
	}

	return db, nil
}

func initServices(_ context.Context, cnf *viper.Viper, db *sql.DB, s3 *minio.Client) (transport.Services, error) {
	return transport.Services{
		User: user.NewService(
			userStorage.NewDatabaseRepository(db),
			user.Options{
				TokenSecret:           []byte(cnf.GetString("TOKEN_SECRET")),
				PasswordSalt:          cnf.GetString("PASSWORD_SALT"),
				TokenExpirationPeriod: cnf.GetDuration("TOKEN_EXPIRATION_PERIOD"),
			},
		),
		Sync: sync.New(
			sync2.NewDatabaseMetadataRepository(),
			sync2.NewS3BlobRepository(s3, cnf.GetString("S3_BUCKET")),
			transaction.NewDatabaseTransactionProvider(db),
		),
	}, nil
}

func serve(ctx context.Context, addr string, services transport.Services) {
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
