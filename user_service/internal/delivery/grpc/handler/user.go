package handler

import (
	"context"
	userpb "e-commerce/pb/user"
	"e-commerce/pkg/utils/errgrpc"
	"e-commerce/user_service/internal/domain"
)

type gRPCHandlerImpl struct {
	userpb.UnimplementedUserServiceServer
	uc domain.Usecase
}

func NewGRPCHandler(uc domain.Usecase) userpb.UserServiceServer {
	return &gRPCHandlerImpl{uc: uc}
}

func (h *gRPCHandlerImpl) CreateUserProfile(ctx context.Context, req *userpb.CreateUserProfileRequest) (*userpb.CreateUserProfileResponse, error) {
	err := h.uc.RegisterProfile(ctx, domain.Profile{UserID: req.GetUserId(), Username: req.GetUsername()})
	if err != nil {
		return nil, errgrpc.ReadInternalError(err)
	}

	return &userpb.CreateUserProfileResponse{Success: true, Message: "Data profile berhasil disimpan"}, nil
}

func (h *gRPCHandlerImpl) UpdateUserProfile(ctx context.Context, req *userpb.UpdateUserProfileRequest) (*userpb.UpdateUserProfileResponse, error) {
	err := h.uc.UpdateProfile(ctx, domain.Profile{Username: req.GetUsername(), FullName: req.GetFullName(), PhoneNumber: req.GetPhoneNumber(), AvatarURL: req.GetAvatarUrl()})
	if err != nil {
		return nil, errgrpc.ReadInternalError(err)
	}
	return &userpb.UpdateUserProfileResponse{Success: true, Message: "Data profile berhasil diperbarui"}, nil
}

func (h *gRPCHandlerImpl) CreateUserAddress(ctx context.Context, req *userpb.CreateUserAddressRequest) (*userpb.CreateUserAddressResponse, error) {
	err := h.uc.AddAddress(ctx, domain.Address{UserID: req.GetUserId(), RecipientName: req.GetRecipientName(), Phone: req.GetPhoneNumber(), FullAddress: req.GetFullAddress(), City: req.GetCity(), PostalCode: req.GetPostalCode(), IsDefault: req.IsDefault})
	if err != nil {
		return nil, errgrpc.ReadInternalError(err)
	}

	return &userpb.CreateUserAddressResponse{Success: true, Message: "Data alamat berhasil disimpan"}, nil
}

func (h *gRPCHandlerImpl) UpdateUserAddress(ctx context.Context, req *userpb.UpdateUserAddressRequest) (*userpb.UpdateUserAddressResponse, error) {
	err := h.uc.UpdateAddress(ctx, domain.Address{RecipientName: req.GetRecipientName(), Phone: req.GetPhoneNumber(), FullAddress: req.GetFullAddress(), City: req.GetCity(), PostalCode: req.GetPostalCode(), IsDefault: req.GetIsDefault()})
	if err != nil {
		return nil, errgrpc.ReadInternalError(err)
	}

	return &userpb.UpdateUserAddressResponse{Success: true, Message: "Data alamat berhasil diperbarui"}, nil
}

func (h *gRPCHandlerImpl) GetUserAddress(ctx context.Context, req *userpb.GetUserAddressRequest) (*userpb.GetUserAddressResponse, error) {
	res, err := h.uc.FindAddressByUserID(ctx, req.GetUserId())
	if err != nil {
		return nil, errgrpc.ReadInternalError(err)
	}

	userAddresses := make([]*userpb.Address, 0, 5)
	for _, v := range res {
		userAddress := &userpb.Address{
			AddressId:     v.ID,
			RecipientName: v.RecipientName,
			PhoneNumber:   v.Phone,
			FullAddress:   v.FullAddress,
			City:          v.City,
			PostalCode:    v.PostalCode,
			IsDefault:     v.IsDefault,
		}
		userAddresses = append(userAddresses, userAddress)
	}

	return &userpb.GetUserAddressResponse{UserId: req.GetUserId(), Addresses: userAddresses}, nil
}
