package handler

import (
	"e-commerce/api_gateway/internal/delivery/http/request"
	"e-commerce/api_gateway/pkg/validator"
	authpb "e-commerce/pb/auth"
	"e-commerce/pkg/utils"
	"e-commerce/pkg/utils/errgrpc"
	"net/http"
)

type AuthHandlerImpl struct {
	client authpb.AuthServiceClient
}

func NewAuthHandler(client authpb.AuthServiceClient) *AuthHandlerImpl {
	return &AuthHandlerImpl{client: client}
}

func (h *AuthHandlerImpl) RegisterHandler(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	ctx := r.Context()

	var authReq request.RegisterRequest

	if err := utils.ReadJSON(r, &authReq); err != nil {
		utils.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "Format JSON tidak valid"})
		return
	}

	if errs := validator.ValidateStruct(authReq); errs != nil {
		utils.WriteJSON(w, http.StatusBadRequest, map[string]any{
			"message": "validasi payload gagal",
			"error":   errs,
		})
		return
	}

	grpcResp, err := h.client.Register(ctx, &authpb.RegisterRequest{Email: authReq.Email, Username: authReq.Username, Password: authReq.Password})
	if err != nil {
		errgrpc.WriteGRPCError(w, err)
		return
	}

	utils.WriteJSON(w, http.StatusCreated, map[string]string{
		"user_id": grpcResp.GetUserId(),
		"message": grpcResp.GetMessage(),
	})

}

func (h *AuthHandlerImpl) LoginHandler(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	ctx := r.Context()

	var authReq request.LoginRequest

	if err := utils.ReadJSON(r, &authReq); err != nil {
		utils.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "Format JSON tidak valid"})
		return
	}

	if errs := validator.ValidateStruct(authReq); errs != nil {
		utils.WriteJSON(w, http.StatusBadRequest, map[string]any{
			"message": "validasi payload gagal",
			"error":   errs,
		})
		return
	}

	grpcResp, err := h.client.Login(ctx, &authpb.LoginRequest{Email: authReq.Email, Password: authReq.Password})
	if err != nil {
		errgrpc.WriteGRPCError(w, err)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    grpcResp.GetRefreshToken(),
		Path:     "/",
		MaxAge:   7 * 24 * 60 * 60,
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteLaxMode,
	})

	utils.WriteJSON(w, http.StatusOK, map[string]string{
		"access_token": grpcResp.GetAccessToken(),
	})

}
