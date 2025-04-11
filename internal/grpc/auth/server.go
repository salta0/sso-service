package auth

import (
	"context"
	"errors"
	"sso/internal/services/auth"

	ssov1 "github.com/salta0/sso-protos/gen/go/sso"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Auth interface {
	Login(ctx context.Context, email string, password string, appID int) (token string, err error)
	RegisterNewUser(ctx context.Context, email string, password string) (userID int64, err error)
	IsAdmin(ctx context.Context, userID int64) (bool, error)
}

type serverAPI struct {
	ssov1.UnimplementedAuthServer
	auth Auth
}

func Register(gRPC *grpc.Server, auth Auth) {
	ssov1.RegisterAuthServer(gRPC, &serverAPI{auth: auth})
}

const (
	emptyIntValue = 0
)

func (s *serverAPI) Login(ctx context.Context, req *ssov1.LoginRequest) (*ssov1.LoginResponse, error) {
	if err := validateLogin(req); err != nil {
		return nil, err
	}

	token, err := s.auth.Login(ctx, req.GetEmail(), req.GetPassword(), int(req.GetAppId()))
	if err != nil {
		if errors.Is(err, auth.ErrorInvalidCreds) {
			return nil, status.Error(codes.InvalidArgument, "invalid login or password")
		}

		if errors.Is(err, auth.ErrorInvalidAppID) {
			return nil, status.Error(codes.InvalidArgument, "invalid app id")
		}

		return nil, status.Error(codes.Internal, "something went wrong")
	}

	return &ssov1.LoginResponse{Token: token}, nil
}

func validateLogin(req *ssov1.LoginRequest) error {
	if req.GetEmail() == "" {
		return status.Error(codes.InvalidArgument, "email is required")
	}

	if req.GetPassword() == "" {
		return status.Error(codes.InvalidArgument, "password is required")
	}

	if req.GetAppId() == emptyIntValue {
		return status.Error(codes.InvalidArgument, "app_id is required")
	}

	return nil
}

func (s *serverAPI) Register(ctx context.Context, req *ssov1.RegisterRequest) (*ssov1.RegisterResponse, error) {
	if err := validateRegister(req); err != nil {
		return nil, err
	}
	userID, err := s.auth.RegisterNewUser(ctx, req.GetEmail(), req.GetPassword())
	if err != nil {
		if errors.Is(err, auth.ErrorUserExists) {
			return nil, status.Error(codes.AlreadyExists, "user already registred")
		}

		return nil, status.Error(codes.Internal, "something went wrong")
	}

	return &ssov1.RegisterResponse{Id: userID}, nil
}

func validateRegister(req *ssov1.RegisterRequest) error {
	if req.GetEmail() == "" {
		return status.Error(codes.InvalidArgument, "email is required")
	}

	if req.GetPassword() == "" {
		return status.Error(codes.InvalidArgument, "password is required")
	}

	return nil
}

func (s *serverAPI) IsAdmin(ctx context.Context, req *ssov1.IsAdminRequest) (*ssov1.IsAdminResponse, error) {
	if err := validateIsAdmin(req); err != nil {
		return nil, err
	}

	isAdmin, err := s.auth.IsAdmin(ctx, req.GetUserId())
	if err != nil {
		if errors.Is(err, auth.ErrorInvalidUserID) {
			return nil, status.Error(codes.InvalidArgument, "user not found")
		}
		return nil, status.Error(codes.Internal, "something went wrong")
	}

	return &ssov1.IsAdminResponse{IsAdmin: isAdmin}, nil
}

func validateIsAdmin(req *ssov1.IsAdminRequest) error {
	if req.GetUserId() == emptyIntValue {
		return status.Error(codes.InvalidArgument, "invalid id")
	}

	return nil
}
