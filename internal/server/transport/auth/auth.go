// Package auth provides authentication utilities for the server, including
// middleware functions for extracting and validating authentication tokens.
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

// tokenInfoKey is a private type used as a key for storing token information in the context.
type tokenInfoKey struct{}

// NewAuthFunc creates a new authentication function that extracts and validates
// a bearer token from the context metadata. It uses the provided userService
// to parse and validate the token. If successful, it injects the token information
// into the context for downstream use.
//
// Returns a function that can be used as a gRPC middleware for authentication.
func NewAuthFunc(userService user.Service) func(ctx context.Context) (context.Context, error) {
	return func(ctx context.Context) (context.Context, error) {
		token, err := authInterceptor.AuthFromMD(ctx, "bearer")
		if err != nil {
			return nil, err
		}

		tokenInfo, err := userService.ParseAuthToken(ctx, token)
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

// SetTokenInfo stores the provided token information in the given context.
// This allows downstream handlers to retrieve the token information.
//
// Returns a new context containing the token information.
func SetTokenInfo(ctx context.Context, tokenInfo user.TokenInfo) context.Context {
	return context.WithValue(ctx, tokenInfoKey{}, tokenInfo)
}

// GetTokenInfo retrieves the token information from the given context.
// If no token information is found, it returns false as the second return value.
//
// Returns the token information and a boolean indicating whether it was found.
func GetTokenInfo(ctx context.Context) (user.TokenInfo, bool) {
	tokenInfo, ok := ctx.Value(tokenInfoKey{}).(user.TokenInfo)

	return tokenInfo, ok
}
