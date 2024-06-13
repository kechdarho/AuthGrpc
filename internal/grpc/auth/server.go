package auth

import (
	"context"
	ssov1 "github.com/idalovkh/authProto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const emptyValue = 0

type Authenticator interface {
	Login(
		ctx context.Context,
		login string,
		password string,
	) (token string, err error)
	RegisterNewUser(
		ctx context.Context,
		email string,
		login string,
		phone string,
		password string,
	) (id int64, err error)
	LogOut(
		ctx context.Context,
		token string,
	) (bool, error)
}

type serverAPI struct {
	ssov1.UnimplementedAuthServer
	auth Authenticator
}

func Register(gRPCServer *grpc.Server, auth Authenticator) {
	ssov1.RegisterAuthServer(gRPCServer, &serverAPI{auth: auth})
}

func (s *serverAPI) Register(ctx context.Context, req *ssov1.RegisterRequest) (*ssov1.RegisterResponse, error) {
	if err := validateRegister(req); err != nil {
		return nil, err
	}

	result, err := s.auth.RegisterNewUser(ctx, req.GetEmail(), req.GetLogin(), req.GetPhone(), req.GetPassword())
	if err != nil {
		// TODO: ...
		return nil, status.Error(codes.Internal, "internal error")
	}

	return &ssov1.RegisterResponse{
		Success: result,
	}, nil
}

func (s *serverAPI) Login(ctx context.Context, req *ssov1.LoginRequest) (*ssov1.LoginResponse, error) {
	if err := validateLogin(req); err != nil {
		return nil, err
	}

	token, err := s.auth.Login(ctx, req.GetLogin(), req.GetPassword())
	if err != nil {
		return nil, status.Error(codes.Internal, "internal error")
	}

	return &ssov1.LoginResponse{
		Token: token,
	}, nil
}

func (s *serverAPI) LogOut(ctx context.Context, req *ssov1.LogoutRequest) (*ssov1.LogoutResponse, error) {
	if err := validateLogOut(req); err != nil {
		return nil, err
	}

	result, err := s.auth.LogOut(ctx, req.GetToken())
	if err != nil {
		return nil, status.Error(codes.Internal, "internal error")
	}

	return &ssov1.LogoutResponse{Success: result}, nil
}

func validateLogin(req *ssov1.LoginRequest) error {
	if req.GetLogin() == "" {
		return status.Error(codes.InvalidArgument, "login")
	}
	if req.GetPassword() == "" {
		return status.Error(codes.InvalidArgument, "Password is required")
	}

	return nil
}

func validateRegister(req *ssov1.RegisterRequest) error {
	if req.GetEmail() == "" {
		return status.Error(codes.InvalidArgument, "email is required")
	}
	if req.GetLogin() == "" {
		return status.Error(codes.InvalidArgument, "login is required")
	}
	if req.GetPassword() == "" {
		return status.Error(codes.InvalidArgument, "Password is required")
	}

	if req.GetLogin() == "" {
		return status.Error(codes.InvalidArgument, "Login is required")
	}

	return nil
}

func validateIsAdmin(req *ssov1.IsAdminRequest) error {
	if req.GetEmail() == "" {
		return status.Error(codes.InvalidArgument, "UserId is required")
	}

	return nil
}

func validateLogOut(req *ssov1.LogoutRequest) error {
	if req.GetToken() == "" {
		return status.Error(codes.InvalidArgument, "Token is required")
	}

	return nil
}
