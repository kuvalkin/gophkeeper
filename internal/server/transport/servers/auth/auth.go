package auth

import (
	"context"
	"errors"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/kuvalkin/gophkeeper/internal/server/service/user"
	pb "github.com/kuvalkin/gophkeeper/pkg/proto/auth/v1"
)

func New(userService user.Service) pb.AuthServiceServer {
	return &server{
		userService: userService,
	}
}

type server struct {
	pb.UnimplementedAuthServiceServer
	userService user.Service
}

func (s *server) AuthFuncOverride(ctx context.Context, _ string) (context.Context, error) {
	// auth service is available without authorization
	return ctx, nil
}

func (s *server) Register(ctx context.Context, request *pb.RegisterRequest) (*pb.RegisterResponse, error) {
	err := s.userService.RegisterUser(ctx, request.Login, request.Password)
	if err != nil {
		if errors.Is(err, user.ErrLoginTaken) {
			return nil, status.Error(codes.AlreadyExists, err.Error())
		}

		if errors.Is(err, user.ErrInvalidLogin) {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}

		return nil, status.Error(codes.Internal, "internal error")
	}

	token, err := s.login(ctx, request.Login, request.Password)
	if err != nil {
		return nil, err
	}

	return &pb.RegisterResponse{Token: token}, nil
}

func (s *server) Login(ctx context.Context, request *pb.LoginRequest) (*pb.LoginResponse, error) {
	token, err := s.login(ctx, request.Login, request.Password)
	if err != nil {
		return nil, err
	}

	return &pb.LoginResponse{Token: token}, nil
}

func (s *server) login(ctx context.Context, login string, password string) (string, error) {
	token, err := s.userService.LoginUser(ctx, login, password)
	if err != nil {
		if errors.Is(err, user.ErrInvalidPair) {
			return "", status.Error(codes.Unauthenticated, err.Error())
		}

		return "", status.Error(codes.Internal, "internal error")
	}

	return token, nil
}
