package auth

import (
	"context"
	"errors"

	authInterceptor "github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/auth"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/kuvalkin/gophkeeper/internal/server/service/user"
)

type tokenInfoKey struct{}

func NewAuthFunc(userService user.Service) func(ctx context.Context) (context.Context, error) {
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

		return SetTokenInfo(ctx, *tokenInfo), nil
	}
}

func SetTokenInfo(ctx context.Context, tokenInfo user.TokenInfo) context.Context {
	return context.WithValue(ctx, tokenInfoKey{}, tokenInfo)
}

func GetTokenInfo(ctx context.Context) (user.TokenInfo, bool) {
	tokenInfo, ok := ctx.Value(tokenInfoKey{}).(user.TokenInfo)

	return tokenInfo, ok
}
