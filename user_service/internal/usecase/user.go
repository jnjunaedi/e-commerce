package usecase

import (
	"context"
	"e-commerce/pkg/utils"
	"e-commerce/pkg/utils/apperror"
	"e-commerce/user_service/internal/domain"
)

type UsecaseImpl struct {
	repo        domain.Repository
	idGenerator utils.IDGenerator
}

func NewUsecase(repo domain.Repository, idGenerator utils.IDGenerator) domain.Usecase {
	return &UsecaseImpl{repo: repo, idGenerator: idGenerator}
}

func (u *UsecaseImpl) RegisterProfile(ctx context.Context, userProfile domain.Profile) error {
	genUUID, err := u.idGenerator.GenerateUUID()
	if err != nil {
		return apperror.New(apperror.ErrInternalServer, "error internal server")
	}
	userProfile.ID = genUUID
	return u.repo.CreateProfile(ctx, userProfile)
}

func (u *UsecaseImpl) UpdateProfile(ctx context.Context, profile domain.Profile) error {
	return u.repo.UpdateProfile(ctx, profile)
}

func (u *UsecaseImpl) AddAddress(ctx context.Context, userAddress domain.Address) error {
	genUUID, err := u.idGenerator.GenerateUUID()
	if err != nil {
		return apperror.New(apperror.ErrInternalServer, "error internal server")
	}
	userAddress.ID = genUUID

	addresses, err := u.repo.GetAddressByUserID(ctx, userAddress.UserID)
	if err == nil && len(addresses) == 0 {
		userAddress.IsDefault = true
	}

	if err != nil {
		return err
	}

	return u.repo.CreateAddress(ctx, userAddress)
}

func (u *UsecaseImpl) UpdateAddress(ctx context.Context, address domain.Address) error {
	return u.repo.UpdateAddress(ctx, address)
}

func (u *UsecaseImpl) FindAddressByUserID(ctx context.Context, userID string) ([]domain.Address, error) {
	return u.repo.GetAddressByUserID(ctx, userID)
}
