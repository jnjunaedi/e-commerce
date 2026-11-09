package usecase

import (
	"context"
	"e-commerce/auth_service/internal/domain"
	"e-commerce/auth_service/pkg/security"
	"e-commerce/pkg/utils"
	"e-commerce/pkg/utils/apperror"
	"encoding/json"
)

type UsecaseImpl struct {
	repo        domain.Repository
	txManager   domain.TxManager
	password    security.Password
	token       security.TokenProvider
	idGenerator utils.IDGenerator
}

func NewUsecase(repo domain.Repository, txManager domain.TxManager, password security.Password, token security.TokenProvider, idGenerator utils.IDGenerator) domain.Usecase {
	return &UsecaseImpl{repo: repo, txManager: txManager, password: password, token: token, idGenerator: idGenerator}
}

func (u *UsecaseImpl) Register(ctx context.Context, req domain.RegisterRequest) (string, error) {
	isExists, err := u.repo.ExistsByEmailOrUsername(ctx, req.Email, req.Username)
	if err != nil {
		return "", err
	}
	if isExists {
		return "", apperror.New(apperror.ErrBadRequest, "email atau username sudah digunakan, silahkan gunakan yang lain")
	}

	hashPw, err := u.password.HashPassword(req.Password)
	if err != nil {
		return "", err
	}

	userID, err := u.idGenerator.GenerateUUID()
	if err != nil {
		return "", err
	}

	newAuth := domain.Auth{
		ID:           userID,
		Email:        req.Email,
		Username:     req.Username,
		PasswordHash: hashPw,
	}

	err = u.txManager.WithTransaction(ctx, func(txCtx context.Context) error {
		err = u.repo.Create(txCtx, newAuth)
		if err != nil {
			return err
		}

		profileData := map[string]string{
			"user_id":  userID,
			"username": req.Username,
		}

		payloadBytes, err := json.Marshal(profileData)
		if err != nil {
			return err
		}

		authboxID, err := u.idGenerator.GenerateUUID()
		if err != nil {
			return err
		}

		err = u.repo.CreateOutbox(txCtx, authboxID, "UserRegistered", payloadBytes)
		if err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return "", err
	}

	return userID, nil
}

func (u *UsecaseImpl) Login(ctx context.Context, req domain.LoginRequest) (domain.TokenResponse, error) {
	authRes, err := u.repo.FindByEmail(ctx, req.Email)
	if err != nil {
		return domain.TokenResponse{}, err
	}

	isMatch := u.password.ComparePassword(req.Password, authRes.PasswordHash)
	if !isMatch {
		return domain.TokenResponse{}, apperror.New(apperror.ErrUnauthorized, "password yang anda masukkan salah")
	}

	userID := authRes.ID

	accessToken, err := u.token.GenerateAccessToken(userID, authRes.Email, authRes.Username)
	if err != nil {
		return domain.TokenResponse{}, err
	}

	refreshToken, err := u.token.GenerateRefreshAccessToken(userID)
	if err != nil {
		return domain.TokenResponse{}, err
	}

	return domain.TokenResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil

}

func (u *UsecaseImpl) ValidateToken(ctx context.Context, token string) (domain.Auth, error) {
	claims, err := u.token.VerifyAccessToken(token)
	if err != nil {
		return domain.Auth{}, err
	}

	return domain.Auth{
		ID:       claims.UserID,
		Email:    claims.Email,
		Username: claims.Username,
	}, nil
}
