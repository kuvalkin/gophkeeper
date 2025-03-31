// Package transport provides the gRPC server setup and configuration for the Gophkeeper application.
// It includes the initialization of services, middleware, and server interceptors.
package transport

import (
	"context"
	"fmt"

	"github.com/bufbuild/protovalidate-go"
	authInterceptor "github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/auth"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"
	protovalidateInterceptor "github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/protovalidate"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/recovery"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/kuvalkin/gophkeeper/internal/server/service/entry"
	"github.com/kuvalkin/gophkeeper/internal/server/service/user"
	"github.com/kuvalkin/gophkeeper/internal/server/transport/auth"
	authServer "github.com/kuvalkin/gophkeeper/internal/server/transport/servers/auth"
	entryServer "github.com/kuvalkin/gophkeeper/internal/server/transport/servers/entry"
	"github.com/kuvalkin/gophkeeper/internal/support/log"
	authpb "github.com/kuvalkin/gophkeeper/pkg/proto/auth/v1"
	entypb "github.com/kuvalkin/gophkeeper/pkg/proto/entry/v1"
)

// Services encapsulates the user and entry services required by the gRPC server.
type Services struct {
	User  user.Service  // User service for handling user-related operations.
	Entry entry.Service // Entry service for handling entry-related operations.
}

// NewServer initializes and returns a new gRPC server configured with the provided services and chunk size.
// It sets up middleware for logging, authentication, validation, and recovery.
//
// Parameters:
//   - services: The Services struct containing the user and entry services.
//   - chunkSize: The size of chunks for entry service operations.
//
// Returns:
//   - A pointer to the configured gRPC server.
//   - An error if the server initialization fails.
func NewServer(services Services, chunkSize int64) (*grpc.Server, error) {
	grpcLog := log.Logger().Named("grpc")

	interceptorLogger := newLogger(grpcLog)
	logOptions := []logging.Option{
		logging.WithLogOnEvents(logging.StartCall, logging.FinishCall),
	}

	authFunc := auth.NewAuthFunc(services.User)

	validator, err := protovalidate.New()
	if err != nil {
		return nil, fmt.Errorf("cant create validator: %w", err)
	}

	recovererFunc := func(p any) error {
		grpcLog.Errorw("panic recovered", "panic", p)

		return status.Error(codes.Internal, "internal server error")
	}

	srv := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			protovalidateInterceptor.UnaryServerInterceptor(validator),
			authInterceptor.UnaryServerInterceptor(authFunc),
			logging.UnaryServerInterceptor(interceptorLogger, logOptions...),
			recovery.UnaryServerInterceptor(recovery.WithRecoveryHandler(recovererFunc)),
		),
		grpc.ChainStreamInterceptor(
			protovalidateInterceptor.StreamServerInterceptor(validator),
			authInterceptor.StreamServerInterceptor(authFunc),
			logging.StreamServerInterceptor(interceptorLogger, logOptions...),
			recovery.StreamServerInterceptor(recovery.WithRecoveryHandler(recovererFunc)),
		),
	)

	authpb.RegisterAuthServiceServer(srv, authServer.New(services.User))
	entypb.RegisterEntryServiceServer(srv, entryServer.New(services.Entry, chunkSize))

	return srv, nil
}

func newLogger(l *zap.SugaredLogger) logging.Logger {
	return logging.LoggerFunc(func(ctx context.Context, lvl logging.Level, msg string, fields ...any) {
		switch lvl {
		case logging.LevelDebug:
			l.Debugw(msg, fields...)
		case logging.LevelInfo:
			l.Infow(msg, fields...)
		case logging.LevelWarn:
			l.Warnw(msg, fields...)
		case logging.LevelError:
			l.Errorw(msg, fields...)
		default:
			panic(fmt.Sprintf("unknown level %v", lvl))
		}
	})
}
