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
	"github.com/kuvalkin/gophkeeper/internal/client/service/secret"
	keyringStorage "github.com/kuvalkin/gophkeeper/internal/client/storage/keyring"
	"github.com/kuvalkin/gophkeeper/internal/client/support/crypt"
	"github.com/kuvalkin/gophkeeper/internal/client/tui/prompts"
	authpb "github.com/kuvalkin/gophkeeper/internal/proto/auth/v1"
	entypb "github.com/kuvalkin/gophkeeper/internal/proto/entry/v1"
	"github.com/kuvalkin/gophkeeper/internal/storage/blob"
)

type Container interface {
	io.Closer
	GetSecretService(ctx context.Context) (secret.Service, error)
	GetAuthService(ctx context.Context) (auth.Service, error)
	GetEntryService(ctx context.Context) (entry.Service, error)
	GetPrompter(ctx context.Context) (prompts.Prompter, error)
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

	initSecretService sync.Once
	secretService     secret.Service

	initAuthService sync.Once
	authService     auth.Service

	initEntryService sync.Once
	entryService     entry.Service

	initCrypter sync.Once
	crypter     *crypt.AgeCrypter

	initPrompter sync.Once
	prompter prompts.Prompter

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

func (c *container) GetPrompter(ctx context.Context) (prompts.Prompter, error) {
	c.initPrompter.Do(func() {
		c.prompter = prompts.NewTerminalPrompter()
	})
	
	return c.prompter, nil
}

func (c *container) GetSecretService(_ context.Context) (secret.Service, error) {
	c.initSecretService.Do(func() {
		c.secretService = secret.New(
			keyringStorage.NewRepository("secret"),
		)
	})

	return c.secretService, nil
}

func (c *container) GetEntryService(ctx context.Context) (entry.Service, error) {
	var outErr error //todo will there be error on second pass?

	c.initEntryService.Do(func() {
		conn, err := c.getConnection()
		if err != nil {
			outErr = fmt.Errorf("cant get grpc connection: %w", err)
			return
		}

		crypter, err := c.getCrypter(ctx)
		if err != nil {
			outErr = fmt.Errorf("cant create crypter: %w", err)
			return
		}

		c.tempDir, err = os.MkdirTemp("", "gophkeeper-*")
		if err != nil {
			outErr = fmt.Errorf("cant create temp dir: %w", err)
			return
		}

		c.entryService = entry.New(
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
			keyringStorage.NewRepository("token"),
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

func (c *container) getCrypter(ctx context.Context) (*crypt.AgeCrypter, error) {
	var outErr error

	c.initCrypter.Do(func() {
		ss, err := c.GetSecretService(ctx)
		if err != nil {
			outErr = fmt.Errorf("cant get secret service: %w", err)
			return
		}

		s, ok, err := ss.Get(ctx)
		if err != nil {
			outErr = fmt.Errorf("cant get secret from keyring: %w", err)
			return
		}

		if !ok {
			outErr = errors.New("secret not found")
			return
		}

		c.crypter, err = crypt.NewAgeCrypter(s)
		if err != nil {
			outErr = fmt.Errorf("cant create crypter: %w", err)
			return
		}
	})

	return c.crypter, outErr
}
