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

	pbAuth "github.com/kuvalkin/gophkeeper/internal/proto/auth/v1"
	pbSync "github.com/kuvalkin/gophkeeper/internal/proto/sync/v1"
	"github.com/kuvalkin/gophkeeper/internal/server/service/user"
	auth2 "github.com/kuvalkin/gophkeeper/internal/server/transport/auth"
	"github.com/kuvalkin/gophkeeper/internal/server/transport/servers/auth"
	"github.com/kuvalkin/gophkeeper/internal/server/transport/servers/sync"
	"github.com/kuvalkin/gophkeeper/internal/support/log"
)

type Services struct {
	User user.Service
}

func NewServer(services Services) (*grpc.Server, error) {
	interceptorLogger := newLogger(log.Logger())
	logOptions := []logging.Option{
		logging.WithLogOnEvents(logging.StartCall, logging.FinishCall),
	}

	authFunc := auth2.NewAuthFunc(services.User)

	validator, err := protovalidate.New()
	if err != nil {
		return nil, fmt.Errorf("cant create validator: %w", err)
	}

	srv := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			protovalidateInterceptor.UnaryServerInterceptor(validator),
			authInterceptor.UnaryServerInterceptor(authFunc),
			logging.UnaryServerInterceptor(interceptorLogger, logOptions...),
			recovery.UnaryServerInterceptor(),
		),
		grpc.ChainStreamInterceptor(
			protovalidateInterceptor.StreamServerInterceptor(validator),
			authInterceptor.StreamServerInterceptor(authFunc),
			logging.StreamServerInterceptor(interceptorLogger, logOptions...),
			recovery.StreamServerInterceptor(),
		),
	)

	pbAuth.RegisterAuthServiceServer(srv, auth.New(services.User))
	pbSync.RegisterSyncServiceServer(srv, sync.New())

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
