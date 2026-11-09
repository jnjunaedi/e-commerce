package handler

import (
	"e-commerce/api_gateway/internal/delivery/http/request"
	"e-commerce/api_gateway/pkg/validator"
	userpb "e-commerce/pb/user"
	"e-commerce/pkg/utils"
	"e-commerce/pkg/utils/errgrpc"
	"net/http"
)

type UserHandlerImpl struct {
	client userpb.UserServiceClient
}

func NewUserHandler(client userpb.UserServiceClient) *UserHandlerImpl {
	return &UserHandlerImpl{client: client}
}

func (h *UserHandlerImpl) UpdateUserProfileHandler(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	ctx := r.Context()

	var userProfile request.UserProfile

	if err := utils.ReadJSON(r, &userProfile); err != nil {
		utils.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "Format JSON tidak valid"})
		return
	}

	if errs := validator.ValidateStruct(userProfile); errs != nil {
		utils.WriteJSON(w, http.StatusBadRequest, map[string]any{
			"message": "validasi payload gagal",
			"error":   errs,
		})
		return
	}

	res, err := h.client.UpdateUserProfile(ctx, &userpb.UpdateUserProfileRequest{UserId: userProfile.UserID, Username: userProfile.Username, FullName: userProfile.FullName, PhoneNumber: userProfile.PhoneNumber, AvatarUrl: userProfile.AvatarURL})
	if err != nil {
		errgrpc.WriteGRPCError(w, err)
		return
	}

	utils.WriteJSON(w, http.StatusOK, map[string]any{
		"status":  res.GetSuccess(),
		"message": res.GetMessage(),
	})
}

func (h *UserHandlerImpl) UpdateUserAddressHandler(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	ctx := r.Context()

	var userAddress request.UserAddress

	if err := utils.ReadJSON(r, &userAddress); err != nil {
		utils.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "Format JSON tidak valid"})
		return
	}

	if errs := validator.ValidateStruct(userAddress); errs != nil {
		utils.WriteJSON(w, http.StatusBadRequest, map[string]any{
			"message": "validasi payload gagal",
			"error":   errs,
		})
		return
	}

	res, err := h.client.UpdateUserAddress(ctx, &userpb.UpdateUserAddressRequest{UserId: userAddress.UserID, RecipientName: userAddress.RecipientName, PhoneNumber: userAddress.PhoneNumber, FullAddress: userAddress.FullAddress, City: userAddress.City, PostalCode: userAddress.PostalCode, IsDefault: userAddress.IsDefault})
	if err != nil {
		errgrpc.WriteGRPCError(w, err)
		return
	}

	utils.WriteJSON(w, http.StatusOK, map[string]any{
		"status":  res.GetSuccess(),
		"message": res.GetMessage(),
	})
}
