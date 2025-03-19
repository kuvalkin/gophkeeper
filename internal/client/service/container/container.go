package container

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"sync"

	"github.com/spf13/viper"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/kuvalkin/gophkeeper/internal/client/cmd"
	"github.com/kuvalkin/gophkeeper/internal/client/service/auth"
	entryService "github.com/kuvalkin/gophkeeper/internal/client/service/entry"
	authStorage "github.com/kuvalkin/gophkeeper/internal/client/storage/auth"
	entryStorage "github.com/kuvalkin/gophkeeper/internal/client/storage/entry"
	"github.com/kuvalkin/gophkeeper/internal/client/support/crypt"
	"github.com/kuvalkin/gophkeeper/internal/client/support/database"
	"github.com/kuvalkin/gophkeeper/internal/client/support/keyring"
	pbAuth "github.com/kuvalkin/gophkeeper/internal/proto/auth/v1"
	pbSync "github.com/kuvalkin/gophkeeper/internal/proto/sync/v1"
)

func New(conf *viper.Viper) (*Container, error) {
	c := &Container{
		conf: conf,
	}

	return c, nil
}

type Container struct {
	conf *viper.Viper

	initConnection sync.Once
	connection     *grpc.ClientConn

	initEntryService sync.Once
	entryService     *entryService.Service

	initAuthService sync.Once
	authService     *auth.Service

	initDB sync.Once
	db     *sql.DB

	initCrypter sync.Once
	crypter     *crypt.AgeCrypter
}

// not goroutine safe!
func (c *Container) Close() error {
	errs := make([]error, 0)
	if c.connection != nil {
		err := c.connection.Close()
		if err != nil {
			errs = append(errs, fmt.Errorf("error closing grpc connection: %w", err))
		}
	}

	return errors.Join(errs...)
}

func (c *Container) GetEntryService(ctx context.Context) (cmd.EntryService, error) {
	var outErr error //todo will there be error on second pass?

	c.initEntryService.Do(func() {
		db, err := c.getDB(ctx)
		if err != nil {
			outErr = fmt.Errorf("cant get db: %w", err)
			return
		}

		metadataRepo, err := entryStorage.NewDatabaseMetadataRepository(db)
		if err != nil {
			outErr = fmt.Errorf("cant create metadata repository: %w", err)
			return
		}

		blobRepo, err := entryStorage.NewFileBlobRepository(c.conf.GetString("storage.blob.path"))
		if err != nil {
			outErr = fmt.Errorf("cant create blob repository: %w", err)
			return
		}

		conn, err := c.getConnection()
		if err != nil {
			outErr = fmt.Errorf("cant get grpc connection: %w", err)
			return
		}

		crypter, err := c.getCrypter()
		if err != nil {
			outErr = fmt.Errorf("cant create crypter: %w", err)
			return
		}

		c.entryService, outErr = entryService.New(
			crypter,
			pbSync.NewSyncServiceClient(conn),
			metadataRepo,
			blobRepo,
		)
	})

	return c.entryService, outErr
}

func (c *Container) GetAuthService(ctx context.Context) (cmd.RegisterService, error) {
	var outErr error

	c.initAuthService.Do(func() {
		db, err := c.getDB(ctx)
		if err != nil {
			outErr = fmt.Errorf("cant get db: %w", err)
			return
		}

		repo, err := authStorage.NewDatabaseRepository(db)
		if err != nil {
			outErr = fmt.Errorf("cant create auth repository: %w", err)
			return
		}

		conn, err := c.getConnection()
		if err != nil {
			outErr = fmt.Errorf("cant get grpc connection: %w", err)
			return
		}

		crypter, err := c.getCrypter()
		if err != nil {
			outErr = fmt.Errorf("cant get crypter: %w", err)
			return
		}

		c.authService, outErr = auth.New(
			pbAuth.NewAuthServiceClient(conn),
			repo,
			crypter,
		)
	})

	return c.authService, outErr
}

func (c *Container) getConnection() (*grpc.ClientConn, error) {
	var err error

	c.initConnection.Do(func() {
		c.connection, err = c.newConnection()
	})

	return c.connection, err
}

func (c *Container) newConnection() (*grpc.ClientConn, error) {
	var creds credentials.TransportCredentials
	if c.conf.GetBool("server.insecure") {
		creds = insecure.NewCredentials()
	} else {
		creds = credentials.NewTLS(nil)
	}

	conn, err := grpc.NewClient(
		c.conf.GetString("server.address"),
		grpc.WithTransportCredentials(creds),
	)
	if err != nil {
		return nil, fmt.Errorf("cant create a grpc client connection: %w", err)
	}

	return conn, nil
}

func (c *Container) getDB(ctx context.Context) (*sql.DB, error) {
	var err error

	c.initDB.Do(func() {
		c.db, err = database.InitDB(ctx, c.conf.GetString("storage.sqlite.path"))
		if err != nil {
			err = fmt.Errorf("cant init database: %w", err)

			return
		}

		err = database.Migrate(ctx, c.db)
		if err != nil {
			err = fmt.Errorf("cant migrate database: %w", err)

			return
		}
	})

	return c.db, err
}

func (c *Container) getCrypter() (*crypt.AgeCrypter, error) {
	var outErr error

	c.initCrypter.Do(func() {
		secret, ok, err := keyring.Get("secret")
		if err != nil {
			outErr = fmt.Errorf("cant get secret from keyring: %w", err)
			return
		}

		if !ok {
			outErr = cmd.ErrNoSecret
			return
		}

		c.crypter, err = crypt.NewAgeCrypter(secret)
		if err != nil {
			outErr = fmt.Errorf("cant create crypter: %w", err)
			return
		}
	})

	return c.crypter, outErr
}
