package secret

import (
	"context"
	"errors"

	"go.uber.org/zap"

	"github.com/kuvalkin/gophkeeper/internal/support/log"
)

type Service interface {
	Set(ctx context.Context, secret string) error
	Get(ctx context.Context) (string, bool, error)
}

type Repository interface {
	Get(ctx context.Context) (string, bool, error)
	Set(ctx context.Context, secret string) error
}

func New(repo Repository) Service {
	return &service{
		repo: repo,
		log:  log.Logger().Named("service.secret"),
	}
}

type service struct {
	repo Repository
	log  *zap.SugaredLogger
}

var ErrInternal = errors.New("internal error")
var ErrAlreadySet = errors.New("secret already set")

func (s *service) Set(ctx context.Context, secret string) error {
	_, exists, err := s.repo.Get(ctx)
	if err != nil {
		s.log.Errorw("error getting secret", "error", err)

		return ErrInternal
	}

	if exists {
		return ErrAlreadySet
	}

	err = s.repo.Set(ctx, secret)
	if err != nil {
		s.log.Errorw("error setting secret", "error", err)

		return ErrInternal
	}

	return nil
}

func (s *service) Get(ctx context.Context) (string, bool, error) {
	secret, exists, err := s.repo.Get(ctx)
	if err != nil {
		s.log.Errorw("error getting secret", "error", err)

		return "", false, ErrInternal
	}

	return secret, exists, nil
}
