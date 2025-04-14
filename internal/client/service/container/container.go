// Package container provides a service container implementation for the GophKeeper client.
// It manages the initialization and lifecycle of various services such as authentication,
// secret management, and data entry handling. The container ensures that resources are
// properly initialized and cleaned up, and provides a unified interface for accessing
// these services.
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
	"github.com/kuvalkin/gophkeeper/internal/storage/blob"
	authpb "github.com/kuvalkin/gophkeeper/pkg/proto/auth/v1"
	entypb "github.com/kuvalkin/gophkeeper/pkg/proto/entry/v1"
)

// Container defines the interface for the application's service container.
// It provides methods to retrieve various services and resources.
type Container interface {
	io.Closer
	// GetSecretService retrieves the secret service, which manages secrets.
	GetSecretService(ctx context.Context) (secret.Service, error)
	// GetAuthService retrieves the authentication service, which handles user authentication.
	GetAuthService(ctx context.Context) (auth.Service, error)
	// GetEntryService retrieves the entry service, which manages data entries.
	GetEntryService(ctx context.Context) (entry.Service, error)
	// GetPrompter retrieves the prompter, which handles user interactions via the terminal.
	GetPrompter(ctx context.Context) (prompts.Prompter, error)
}

// New creates a new instance of the service container.
// It takes a viper configuration object as input and returns a Container instance or an error.
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
	prompter     prompts.Prompter

	tempDir string
}

// Close releases all resources held by the container, such as temporary directories
// and gRPC connections. It is not goroutine-safe.
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

// GetPrompter initializes and retrieves the terminal prompter for user interactions.
func (c *container) GetPrompter(ctx context.Context) (prompts.Prompter, error) {
	c.initPrompter.Do(func() {
		c.prompter = prompts.NewTerminalPrompter()
	})

	return c.prompter, nil
}

// GetSecretService initializes and retrieves the secret service, which manages secrets
// using a keyring-based storage repository.
func (c *container) GetSecretService(_ context.Context) (secret.Service, error) {
	c.initSecretService.Do(func() {
		c.secretService = secret.New(
			keyringStorage.NewRepository(),
		)
	})

	return c.secretService, nil
}

// GetEntryService initializes and retrieves the entry service, which manages data entries.
// It sets up a gRPC connection, a crypter for encryption, and a temporary directory for file storage.
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

		br, err := blob.NewFileBlobRepository(c.tempDir)
		if err != nil {
			outErr = fmt.Errorf("cant create blob repository: %w", err)
			return
		}

		c.entryService = entry.New(
			crypter,
			entypb.NewEntryServiceClient(conn),
			br,
			c.conf.GetInt64("stream.chunk_size"),
		)
	})

	return c.entryService, outErr
}

// GetAuthService initializes and retrieves the authentication service, which handles user authentication
// using a gRPC connection and a keyring-based storage repository.
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
			keyringStorage.NewRepository(),
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

		s, ok, err := ss.GetSecret(ctx)
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
