package user

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"go.uber.org/zap"

	"github.com/kuvalkin/gophkeeper/internal/support/log"
)

var signingMethod = jwt.SigningMethodHS256

// NewService creates a new instance of the user service with the given repository and options.
func NewService(repo Repository, options Options) Service {
	return &service{
		repo:    repo,
		options: options,
		logger:  log.Logger().Named("userService"),
	}
}

type service struct {
	repo    Repository
	options Options
	logger  *zap.SugaredLogger
}

func (s *service) RegisterUser(ctx context.Context, login string, password string) error {
	if login == "" {
		return ErrInvalidLogin
	}

	hash := s.hashPassword(password)

	err := s.repo.AddUser(ctx, login, hash)
	if err != nil {
		if errors.Is(err, ErrLoginNotUnique) {
			return ErrLoginTaken
		}

		s.logger.Errorw("user adding failed", "error", err)

		return ErrInternal
	}

	return nil
}

func (s *service) LoginUser(ctx context.Context, login string, password string) (string, error) {
	userInfo, found, err := s.repo.FindUser(ctx, login)
	if err != nil {
		s.logger.Errorw("failed to fetch password hash", "login", login, "error", err)

		return "", ErrInternal
	}

	if !found {
		return "", ErrInvalidPair
	}

	if userInfo.PasswordHash != s.hashPassword(password) {
		return "", ErrInvalidPair
	}

	token, err := s.issueToken(userInfo.ID)
	if err != nil {
		s.logger.Errorw("failed to issue token", "login", login, "error", err)

		return "", ErrInternal
	}

	return token, nil
}

func (s *service) ParseAuthToken(_ context.Context, token string) (*TokenInfo, error) {
	claims := new(jwt.RegisteredClaims)

	parsedToken, err := jwt.ParseWithClaims(
		token,
		claims,
		func(t *jwt.Token) (interface{}, error) {
			method, ok := t.Method.(*jwt.SigningMethodHMAC)
			if !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
			}

			if method.Name != signingMethod.Name {
				return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
			}

			return s.options.TokenSecret, nil
		},
	)

	if err != nil {
		s.logger.Infow("failed to parse token", "error", err)

		return nil, ErrInvalidToken
	}

	if !parsedToken.Valid {
		return nil, ErrInvalidToken
	}

	return &TokenInfo{UserID: claims.Subject}, nil
}

func (s *service) hashPassword(password string) string {
	withSalt := password + s.options.PasswordSalt

	hashBytes := sha256.Sum256([]byte(withSalt))

	return hex.EncodeToString(hashBytes[:])
}

func (s *service) issueToken(userID string) (string, error) {
	now := time.Now()

	token := jwt.NewWithClaims(signingMethod, jwt.RegisteredClaims{
		Subject:   userID,
		IssuedAt:  jwt.NewNumericDate(now),
		ExpiresAt: jwt.NewNumericDate(now.Add(s.options.TokenExpirationPeriod)),
	})

	tokenString, err := token.SignedString(s.options.TokenSecret)
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %w", err)
	}

	return tokenString, nil
}
