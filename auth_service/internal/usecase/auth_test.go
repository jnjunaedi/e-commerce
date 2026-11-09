package usecase_test

import (
	"context"
	"e-commerce/auth_service/internal/domain"
	mRepo "e-commerce/auth_service/internal/repository/mocks"
	mTx "e-commerce/auth_service/internal/repository/mocks"
	"e-commerce/auth_service/internal/usecase"
	"e-commerce/auth_service/pkg/security"
	mPassword "e-commerce/auth_service/pkg/security/mocks"
	mId "e-commerce/pkg/utils/mocks"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestAuthUsecase_Register(t *testing.T) {
	mockRepo := mRepo.NewRepository(t)
	mockPassword := mPassword.NewPassword(t)
	mockToken := mPassword.NewTokenProvider(t)
	mockId := mId.NewIDGenerator(t)
	mockTx := mTx.NewTxManager(t)

	pw := "jujun123"

	RegDummy := domain.RegisterRequest{
		Email:    "jujun@gmail.com",
		Username: "jujunjunaedi",
		Password: pw,
	}

	userDummy := "gen-usr-112"
	authDummy := "auth-112"

	mockRepo.On("ExistsByEmailOrUsername", mock.Anything, RegDummy.Email, RegDummy.Username).Return(false, nil)

	mockPassword.On("HashPassword", mock.Anything).Return(pw, nil)

	mockId.On("GenerateUUID").Return(userDummy, nil).Once()

	mockTx.On("WithTransaction", mock.Anything, mock.Anything).Return(func(ctx context.Context, fn func(ctx context.Context) error) error {
		return fn(ctx)
	}, nil)

	mockRepo.On("Create", mock.Anything, mock.MatchedBy(func(a domain.Auth) bool {
		return a.Email == RegDummy.Email && a.Username == RegDummy.Username && a.PasswordHash == RegDummy.Password
	})).Return(nil)

	mockId.On("GenerateUUID").Return(authDummy, nil).Once()

	mockRepo.On("CreateOutbox", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)

	uc := usecase.NewUsecase(mockRepo, mockTx, mockPassword, mockToken, mockId)
	ctx := context.Background()

	res, err := uc.Register(ctx, RegDummy)

	assert.NoError(t, err)
	assert.NotEmpty(t, res)
	assert.Equal(t, userDummy, res)

	mockRepo.AssertExpectations(t)
	mockToken.AssertExpectations(t)
	mockId.AssertExpectations(t)
	mockTx.AssertExpectations(t)
}

func TestAuthUsecase_Login(t *testing.T) {
	mockRepo := mRepo.NewRepository(t)
	mockPassword := mPassword.NewPassword(t)
	mockToken := mPassword.NewTokenProvider(t)
	mockId := mId.NewIDGenerator(t)
	mockTx := mTx.NewTxManager(t)

	authDummy := domain.Auth{
		ID:           "USR-112",
		Email:        "jujun@gmail.com",
		Username:     "jujunjunaedi",
		PasswordHash: "hashed-password-dummy",
	}
	loginReq := domain.LoginRequest{
		Email:    "jujun@gmail.com",
		Password: "hashed-password-dummy",
	}

	genAccessToken := "mocked-access-token"
	genRefreshToken := "mocked-refresh-token"

	mockRepo.On("FindByEmail", mock.Anything, mock.Anything).Return(authDummy, nil)

	mockPassword.On("ComparePassword", loginReq.Password, authDummy.PasswordHash).Return(true)

	mockToken.On("GenerateAccessToken", mock.Anything, authDummy.Email, authDummy.Username).Return(genAccessToken, nil)

	mockToken.On("GenerateRefreshAccessToken", mock.Anything).Return(genRefreshToken, nil)

	uc := usecase.NewUsecase(mockRepo, mockTx, mockPassword, mockToken, mockId)
	ctx := context.Background()

	res, err := uc.Login(ctx, loginReq)

	assert.NoError(t, err)
	assert.Equal(t, genAccessToken, res.AccessToken)
	assert.Equal(t, genRefreshToken, res.RefreshToken)

	mockRepo.AssertExpectations(t)
	mockToken.AssertExpectations(t)
	mockId.AssertExpectations(t)
	mockTx.AssertExpectations(t)
}

func TestAuthUsecase_ValidateToken(t *testing.T) {
	mockRepo := mRepo.NewRepository(t)
	mockPassword := mPassword.NewPassword(t)
	mockToken := mPassword.NewTokenProvider(t)
	mockId := mId.NewIDGenerator(t)
	mockTx := mTx.NewTxManager(t)

	authDummy := &security.TokenClaims{
		UserID:   "USR-112",
		Email:    "jujun@gmail.com",
		Username: "jujunjunaedi",
	}

	tokenDummy := "token-dummy"

	mockToken.On("VerifyAccessToken", tokenDummy).Return(authDummy, nil)

	uc := usecase.NewUsecase(mockRepo, mockTx, mockPassword, mockToken, mockId)
	ctx := context.Background()

	res, err := uc.ValidateToken(ctx, tokenDummy)

	assert.NoError(t, err)
	assert.Equal(t, authDummy.UserID, res.ID)
	assert.Equal(t, authDummy.Email, res.Email)
	assert.Equal(t, authDummy.Username, res.Username)

	mockRepo.AssertExpectations(t)
	mockToken.AssertExpectations(t)
	mockId.AssertExpectations(t)
	mockTx.AssertExpectations(t)
}
