package transport

import (
	"context"
	"errors"
	"fmt"

	authInterceptor "github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/auth"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/recovery"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	pbAuth "github.com/kuvalkin/gophkeeper/internal/proto/auth/v1"
	pbSync "github.com/kuvalkin/gophkeeper/internal/proto/sync/v1"
	"github.com/kuvalkin/gophkeeper/internal/server/service/user"
	"github.com/kuvalkin/gophkeeper/internal/server/transport/servers/auth"
	"github.com/kuvalkin/gophkeeper/internal/server/transport/servers/sync"
	"github.com/kuvalkin/gophkeeper/internal/support/log"
)

type Services struct {
	User user.Service
}

func NewServer(services Services) *grpc.Server {
	interceptorLogger := newLogger(log.Logger())

	logOptions := []logging.Option{
		logging.WithLogOnEvents(logging.StartCall, logging.FinishCall),
	}

	authFunc := newAuthFunc(services.User)

	srv := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			authInterceptor.UnaryServerInterceptor(authFunc),
			logging.UnaryServerInterceptor(interceptorLogger, logOptions...),
			recovery.UnaryServerInterceptor(),
		),
		grpc.ChainStreamInterceptor(
			authInterceptor.StreamServerInterceptor(authFunc),
			logging.StreamServerInterceptor(interceptorLogger, logOptions...),
			recovery.StreamServerInterceptor(),
		),
	)

	pbAuth.RegisterAuthServiceServer(srv, auth.New(services.User))
	pbSync.RegisterSyncServiceServer(srv, sync.New())

	return srv
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

type tokenInfoKey struct{}

func newAuthFunc(userService user.Service) func(ctx context.Context) (context.Context, error) {
	return func(ctx context.Context) (context.Context, error) {
		token, err := authInterceptor.AuthFromMD(ctx, "bearer")
		if err != nil {
			return nil, err
		}

		tokenInfo, err := userService.ParseToken(ctx, token)
		if err != nil {
			if errors.Is(err, user.ErrInvalidToken) {
				return nil, status.Errorf(codes.Unauthenticated, "invalid auth token: %v", err)
			}

			return nil, status.Error(codes.Internal, "internal error")
		}

		ctx = logging.InjectFields(ctx, logging.Fields{"auth.userID", tokenInfo.UserID})

		return context.WithValue(ctx, tokenInfoKey{}, tokenInfo), nil
	}
}
