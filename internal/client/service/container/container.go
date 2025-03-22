package container

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"sync"

	"github.com/spf13/viper"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/kuvalkin/gophkeeper/internal/client/service/auth"
	"github.com/kuvalkin/gophkeeper/internal/client/service/entry"
	authStorage "github.com/kuvalkin/gophkeeper/internal/client/storage/auth"
	"github.com/kuvalkin/gophkeeper/internal/client/support/crypt"
	"github.com/kuvalkin/gophkeeper/internal/client/support/keyring"
	authpb "github.com/kuvalkin/gophkeeper/internal/proto/auth/v1"
	entypb "github.com/kuvalkin/gophkeeper/internal/proto/entry/v1"
	"github.com/kuvalkin/gophkeeper/internal/storage/blob"
)

type Container interface {
	io.Closer
	GetEntryService(ctx context.Context) (entry.Service, error)
	GetAuthService(ctx context.Context) (auth.Service, error)
}

func New(conf *viper.Viper) (Container, error) {
	c := &container{
		conf: conf,
	}

	return c, nil
}

type container struct {
	conf *viper.Viper

	initConnection sync.Once
	connection     *grpc.ClientConn

	initEntryService sync.Once
	entryService     entry.Service

	initAuthService sync.Once
	authService     auth.Service

	initCrypter sync.Once
	crypter     *crypt.AgeCrypter

	tempDir string
}

// not goroutine safe!
func (c *container) Close() error {
	errs := make([]error, 0)

	if c.tempDir != "" {
		err := os.RemoveAll(c.tempDir)
		if err != nil {
			errs = append(errs, fmt.Errorf("error removing temp dir: %w", err))
		}
	}

	if c.connection != nil {
		err := c.connection.Close()
		if err != nil {
			errs = append(errs, fmt.Errorf("error closing grpc connection: %w", err))
		}
	}

	return errors.Join(errs...)
}

func (c *container) GetEntryService(_ context.Context) (entry.Service, error) {
	var outErr error //todo will there be error on second pass?

	c.initEntryService.Do(func() {
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

		c.tempDir, err = os.MkdirTemp("", "gophkeeper-*")
		if err != nil {
			outErr = fmt.Errorf("cant create temp dir: %w", err)
			return
		}

		c.entryService, outErr = entry.New(
			crypter,
			entypb.NewEntryServiceClient(conn),
			blob.NewFileBlobRepository(c.tempDir),
			c.conf.GetInt64("stream.chunk_size"),
		)
	})

	return c.entryService, outErr
}

func (c *container) GetAuthService(_ context.Context) (auth.Service, error) {
	var outErr error

	c.initAuthService.Do(func() {
		conn, err := c.getConnection()
		if err != nil {
			outErr = fmt.Errorf("cant get grpc connection: %w", err)
			return
		}

		c.authService = auth.New(
			authpb.NewAuthServiceClient(conn),
			authStorage.NewKeyringRepository(),
		)
	})

	return c.authService, outErr
}

func (c *container) getConnection() (*grpc.ClientConn, error) {
	var err error

	c.initConnection.Do(func() {
		c.connection, err = c.newConnection()
	})

	return c.connection, err
}

func (c *container) newConnection() (*grpc.ClientConn, error) {
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

func (c *container) getCrypter() (*crypt.AgeCrypter, error) {
	var outErr error

	c.initCrypter.Do(func() {
		secret, ok, err := keyring.Get("secret")
		if err != nil {
			outErr = fmt.Errorf("cant get secret from keyring: %w", err)
			return
		}

		if !ok {
			outErr = auth.ErrNoSecret
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
