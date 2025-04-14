// Package auth provides the implementation of the authentication service
// for the gophkeeper application. It handles user registration and login
// functionalities using gRPC.
package auth

import (
	"context"
	"errors"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/kuvalkin/gophkeeper/internal/server/service/user"
	pb "github.com/kuvalkin/gophkeeper/pkg/proto/auth/v1"
)

// New creates a new instance of the authentication service server.
// It takes a user service as a dependency to handle user-related operations.
func New(userService user.Service) pb.AuthServiceServer {
	return &server{
		userService: userService,
	}
}

type server struct {
	pb.UnimplementedAuthServiceServer
	userService user.Service
}

// AuthFuncOverride allows the authentication service to be accessed
// without requiring prior authorization. It returns the context as is.
func (s *server) AuthFuncOverride(ctx context.Context, _ string) (context.Context, error) {
	// auth service is available without authorization
	return ctx, nil
}

// Register handles user registration requests. It validates the input,
// registers the user, and returns a token upon successful registration.
// If the login is already taken or invalid, it returns appropriate errors.
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

// Login handles user login requests. It validates the credentials and
// returns a token upon successful authentication. If the credentials
// are invalid, it returns an error.
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
