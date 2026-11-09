package usecase_test

import (
	"context"
	mId "e-commerce/pkg/utils/mocks"
	"e-commerce/user_service/internal/domain"
	"e-commerce/user_service/internal/repository/mocks"
	"e-commerce/user_service/internal/usecase"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestUserUsecase_RegisterProfile(t *testing.T) {
	mockRepo := mocks.NewRepository(t)
	mockId := mId.NewIDGenerator(t)

	profile := domain.Profile{
		UserID:      "USR-112",
		Username:    "jujunjunaedi123",
		FullName:    "jujun junaedi",
		PhoneNumber: "085846430761",
		AvatarURL:   "jujun.png",
	}

	idDummy := "gen-usr-112"

	mockId.On("GenerateUUID").Return(idDummy, nil)

	mockRepo.On("CreateProfile", mock.Anything, mock.MatchedBy(func(p domain.Profile) bool {
		return p.ID == idDummy && p.UserID == profile.UserID && p.FullName == profile.FullName && p.PhoneNumber == profile.PhoneNumber && p.AvatarURL == profile.AvatarURL
	})).Return(nil)

	uc := usecase.NewUsecase(mockRepo, mockId)
	ctx := context.Background()

	err := uc.RegisterProfile(ctx, profile)

	assert.NoError(t, err)

	mockRepo.AssertExpectations(t)
	mockId.AssertExpectations(t)
}

func TestUserUsecase_AddAddress(t *testing.T) {
	mockRepo := mocks.NewRepository(t)
	mockId := mId.NewIDGenerator(t)

	address := domain.Address{
		UserID:        "USR-112",
		RecipientName: "jujun junaedi",
		Phone:         "085846430761",
		FullAddress:   "Gg Maryam, Kp Rancabali kulon, Kel Bojongherang, Kec Cianjur, Jawa barat",
		City:          "Cianjur",
		PostalCode:    "43206",
		IsDefault:     true,
	}

	var addresses = []domain.Address{
		domain.Address{
			UserID:        "USR-112",
			RecipientName: "jujun junaedi",
			Phone:         "085846430761",
			FullAddress:   "Gg Maryam, Kp Rancabali kulon, Kel Bojongherang, Kec Cianjur, Jawa barat",
			City:          "Cianjur",
			PostalCode:    "43206",
			IsDefault:     true,
		},
		domain.Address{
			UserID:        "USR-112",
			RecipientName: "alif bratata",
			Phone:         "087566345121",
			FullAddress:   "Gg Rambutan, Kp Rambutan, Kel Bojongherang, Kec Cianjur, Jawa barat",
			City:          "Cianjur",
			PostalCode:    "43206",
			IsDefault:     false,
		},
	}

	idDummy := "gen-usr-112"

	mockId.On("GenerateUUID").Return(idDummy, nil)

	mockRepo.On("GetAddressByUserID", mock.Anything, mock.Anything).Return(addresses, nil)

	mockRepo.On("CreateAddress", mock.Anything, mock.MatchedBy(func(a domain.Address) bool {
		return a.ID == idDummy && a.UserID == address.UserID && a.RecipientName == address.RecipientName && a.Phone == address.Phone && a.FullAddress == address.FullAddress && a.City == address.City && a.PostalCode == address.PostalCode && a.IsDefault == address.IsDefault
	})).Return(nil)

	uc := usecase.NewUsecase(mockRepo, mockId)
	ctx := context.Background()

	err := uc.AddAddress(ctx, address)

	assert.NoError(t, err)

	mockRepo.AssertExpectations(t)
	mockId.AssertExpectations(t)
}

func TestUserUsecase_UpdateProfile(t *testing.T) {
	mockRepo := mocks.NewRepository(t)
	mockId := mId.NewIDGenerator(t)

	profile := domain.Profile{
		UserID:      "USR-112",
		Username:    "jujunjunaedi123",
		FullName:    "jujun junaedi",
		PhoneNumber: "085846430761",
		AvatarURL:   "jujun.png",
	}

	mockRepo.On("UpdateProfile", mock.Anything, mock.MatchedBy(func(p domain.Profile) bool {
		return p.UserID != "" && p.FullName == profile.FullName && p.PhoneNumber == profile.PhoneNumber && p.AvatarURL == profile.AvatarURL
	})).Return(nil)

	uc := usecase.NewUsecase(mockRepo, mockId)
	ctx := context.Background()

	err := uc.UpdateProfile(ctx, profile)

	assert.NoError(t, err)

	mockRepo.AssertExpectations(t)
	mockId.AssertExpectations(t)
}

func TestUserUsecase_UpdateAddress(t *testing.T) {
	mockRepo := mocks.NewRepository(t)
	mockId := mId.NewIDGenerator(t)

	address := domain.Address{
		UserID:        "USR-112",
		RecipientName: "jujun junaedi",
		Phone:         "085846430761",
		FullAddress:   "Gg Maryam, Kp Rancabali kulon, Kel Bojongherang, Kec Cianjur, Jawa barat",
		City:          "Cianjur",
		PostalCode:    "43206",
		IsDefault:     true,
	}

	mockRepo.On("UpdateAddress", mock.Anything, mock.MatchedBy(func(a domain.Address) bool {
		return a.UserID != "" && a.RecipientName == address.RecipientName && a.Phone == address.Phone && a.FullAddress == address.FullAddress && a.City == address.City && a.PostalCode == address.PostalCode && a.IsDefault == address.IsDefault
	})).Return(nil)

	uc := usecase.NewUsecase(mockRepo, mockId)
	ctx := context.Background()

	err := uc.UpdateAddress(ctx, address)

	assert.NoError(t, err)

	mockRepo.AssertExpectations(t)
	mockId.AssertExpectations(t)
}

func TestUserUsecase_FindAddressByUserID(t *testing.T) {
	mockRepo := mocks.NewRepository(t)
	mockId := mId.NewIDGenerator(t)

	var addresses = []domain.Address{
		domain.Address{
			UserID:        "USR-112",
			RecipientName: "jujun junaedi",
			Phone:         "085846430761",
			FullAddress:   "Gg Maryam, Kp Rancabali kulon, Kel Bojongherang, Kec Cianjur, Jawa barat",
			City:          "Cianjur",
			PostalCode:    "43206",
			IsDefault:     true,
		},
		domain.Address{
			UserID:        "USR-112",
			RecipientName: "alif bratata",
			Phone:         "087566345121",
			FullAddress:   "Gg Rambutan, Kp Rambutan, Kel Bojongherang, Kec Cianjur, Jawa barat",
			City:          "Cianjur",
			PostalCode:    "43206",
			IsDefault:     false,
		},
	}

	mockRepo.On("GetAddressByUserID", mock.Anything, mock.Anything).Return(addresses, nil)

	uc := usecase.NewUsecase(mockRepo, mockId)
	ctx := context.Background()

	res, err := uc.FindAddressByUserID(ctx, "USR-112")

	assert.NoError(t, err)
	assert.Len(t, res, 2)

	mockRepo.AssertExpectations(t)
	mockId.AssertExpectations(t)
}
