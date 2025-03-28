package user

import (
	"context"
	"errors"
	"time"
)

var ErrInvalidLogin = errors.New("invalid login")
var ErrLoginTaken = errors.New("user with this login already exists")
var ErrInvalidPair = errors.New("login/password pair is invalid")
var ErrInvalidToken = errors.New("invalid token")
var ErrInternal = errors.New("internal error")

type TokenInfo struct {
	UserID string
}

type Service interface {
	RegisterUser(ctx context.Context, login string, password string) error
	// LoginUser authenticates a user and returns auth token on success
	LoginUser(ctx context.Context, login string, password string) (string, error)
	ParseAuthToken(ctx context.Context, token string) (*TokenInfo, error)
}

type Options struct {
	TokenSecret           []byte
	PasswordSalt          string
	TokenExpirationPeriod time.Duration
}

var ErrLoginNotUnique = errors.New("user with this login already exists")

type UserInfo struct {
	ID       string
	PasswordHash string
}

type Repository interface {
	AddUser(ctx context.Context, login string, passwordHash string) error
	FindUser(ctx context.Context, login string) (UserInfo, bool, error)
}
