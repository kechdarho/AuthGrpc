package auth

import (
	"context"
	ssov1 "github.com/kechdarho/authproto/gen/go/sso"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

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
	) (success bool, err error)
	LogOut(
		ctx context.Context,
		token string,
	) (success bool, err error)
	ChangePassword(
		ctx context.Context,
		login string,
		oldPassword string,
		newPassword string,
	) (success bool, err error)
	ForgotPassword(
		ctx context.Context,
		login string,
	) (string, error)
	ResetPassword(
		ctx context.Context,
		token string,
		newPassword string,
	) (success bool, err error)
	UpdateUser(tx context.Context,
		jwtToken string,
		email string,
		phone string,
	) (success bool, err error)
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

func (s *serverAPI) LogOut(ctx context.Context, req *ssov1.LogOutRequest) (*ssov1.LogOutResponse, error) {
	if err := validateLogOut(req); err != nil {
		return nil, err
	}

	result, err := s.auth.LogOut(ctx, req.GetToken())
	if err != nil {
		return nil, status.Error(codes.Internal, "internal error")
	}
	return &ssov1.LogOutResponse{Success: result}, nil
}

func (s *serverAPI) ChangePassword(ctx context.Context, req *ssov1.ChangePasswordRequest) (*ssov1.ChangePasswordResponse, error) {
	if err := validateChangePassword(req); err != nil {
		return nil, err
	}
	result, err := s.auth.ChangePassword(ctx, req.GetLogin(), req.GetOldPassword(), req.GetNewPassword())
	if err != nil {
		return nil, status.Error(codes.Internal, "internal error")
	}
	return &ssov1.ChangePasswordResponse{Success: result}, nil
}

func (s *serverAPI) ForgotPassword(ctx context.Context, req *ssov1.ForgotPasswordRequest) (*ssov1.ForgotPasswordResponse, error) {
	if err := validateForgotPassword(req); err != nil {
		return nil, err
	}

	resetToken, err := s.auth.ForgotPassword(ctx, req.GetLogin())
	if err != nil {
		return nil, status.Error(codes.Internal, "internal error")
	}
	return &ssov1.ForgotPasswordResponse{Message: resetToken}, nil
}

func (s *serverAPI) ResetPassword(ctx context.Context, req *ssov1.ResetPasswordRequest) (*ssov1.ResetPasswordResponse, error) {
	if err := validateResetPassword(req); err != nil {
		return nil, err
	}
	result, err := s.auth.ResetPassword(ctx, req.GetToken(), req.GetNewPassword())
	if err != nil {
		return nil, status.Error(codes.Internal, "internal error")
	}
	return &ssov1.ResetPasswordResponse{Success: result}, nil
}

func (s *serverAPI) UpdateUser(ctx context.Context, req *ssov1.UpdateUserRequest) (*ssov1.UpdateUserResponse, error) {
	if err := validateUpdateUser(req); err != nil {
		return nil, err
	}
	result, err := s.auth.UpdateUser(ctx, req.GetToken(), req.GetEmail(), req.GetPhone())
	if err != nil {
		return nil, status.Error(codes.Internal, "internal error")
	}
	return &ssov1.UpdateUserResponse{Success: result}, nil
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

func validateUpdateUser(req *ssov1.UpdateUserRequest) error {
	if req.GetToken() == "" {
		return status.Error(codes.InvalidArgument, "token is required")
	}
	if req.GetEmail() == "" {
		if req.GetPhone() == "" {
			return status.Error(codes.InvalidArgument, "phone or email is required")
		}
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

func validateLogOut(req *ssov1.LogOutRequest) error {
	if req.GetToken() == "" {
		return status.Error(codes.InvalidArgument, "Token is required")
	}

	return nil
}

func validateChangePassword(req *ssov1.ChangePasswordRequest) error {
	if req.GetLogin() == "" {
		return status.Error(codes.InvalidArgument, "login is required")
	}
	if req.GetOldPassword() == "" {
		return status.Error(codes.InvalidArgument, "Old password is required")
	}
	if req.GetNewPassword() == "" {
		return status.Error(codes.InvalidArgument, "New password is required")
	}
	return nil
}

func validateForgotPassword(req *ssov1.ForgotPasswordRequest) error {
	if req.GetLogin() == "" {
		return status.Error(codes.InvalidArgument, "Token is required")
	}

	return nil
}

func validateResetPassword(req *ssov1.ResetPasswordRequest) error {
	if req.GetToken() == "" {
		return status.Error(codes.InvalidArgument, "Token is required")
	}
	if req.GetNewPassword() == "" {
		return status.Error(codes.InvalidArgument, "Token is required")
	}
	return nil
}
