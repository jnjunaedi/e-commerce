package handler

import (
	"context"
	"e-commerce/auth_service/internal/domain"
	authpb "e-commerce/pb/auth"
	"e-commerce/pkg/utils/errgrpc"
)

type gRPCHandlerImpl struct {
	authpb.UnimplementedAuthServiceServer
	uc domain.Usecase
}

func NewGRPCHandler(uc domain.Usecase) authpb.AuthServiceServer {
	return &gRPCHandlerImpl{uc: uc}
}

func (h *gRPCHandlerImpl) Register(ctx context.Context, req *authpb.RegisterRequest) (*authpb.RegisterResponse, error) {
	userID, err := h.uc.Register(ctx, domain.RegisterRequest{Email: req.GetEmail(), Username: req.GetUsername(), Password: req.GetPassword()})
	if err != nil {
		return nil, errgrpc.ReadInternalError(err)
	}
	return &authpb.RegisterResponse{Message: "Registrasi akun berhasil", UserId: userID}, nil
}
func (h *gRPCHandlerImpl) Login(ctx context.Context, req *authpb.LoginRequest) (*authpb.LoginResponse, error) {
	res, err := h.uc.Login(ctx, domain.LoginRequest{Email: req.GetEmail(), Password: req.GetPassword()})
	if err != nil {
		return nil, errgrpc.ReadInternalError(err)
	}

	return &authpb.LoginResponse{AccessToken: res.AccessToken, RefreshToken: res.RefreshToken}, nil
}
func (h *gRPCHandlerImpl) ValidateToken(ctx context.Context, req *authpb.ValidateTokenRequest) (*authpb.ValidateTokenResponse, error) {
	res, err := h.uc.ValidateToken(ctx, req.GetAccessToken())
	if err != nil {
		return &authpb.ValidateTokenResponse{
			IsValid:      false,
			UserId:       res.ID,
			Email:        res.Email,
			Username:     res.Username,
			ErrorMessage: err.Error(),
		}, nil
	}

	return &authpb.ValidateTokenResponse{IsValid: true, UserId: res.ID, Email: res.Email, Username: res.Username}, nil
}
