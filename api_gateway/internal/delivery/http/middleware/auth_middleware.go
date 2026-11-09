package middleware

import (
	"context"
	authpb "e-commerce/pb/auth"
	"e-commerce/pkg/utils"
	"net/http"
	"strings"
	"time"
)

type contextKey string

const UserIDKey contextKey = "user_id"

type AuthMiddleware struct {
	authClient authpb.AuthServiceClient
}

func NewAuthMiddleware(authClient authpb.AuthServiceClient) *AuthMiddleware {
	return &AuthMiddleware{authClient: authClient}
}

func (m *AuthMiddleware) RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			utils.WriteJSON(w, http.StatusUnauthorized, map[string]string{"message": "akses ditolak, silakan login terlebih dahulu"})
			return
		}
		tokenString := strings.TrimPrefix(authHeader, "Bearer ")

		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()

		res, err := m.authClient.ValidateToken(ctx, &authpb.ValidateTokenRequest{AccessToken: tokenString})
		if err != nil && !res.IsValid {
			utils.WriteJSON(w, http.StatusUnauthorized, map[string]string{"message": "token tidak sah atau sudah kedaluwarsa"})
			return
		}

		newCtx := context.WithValue(r.Context(), UserIDKey, res.UserId)

		next.ServeHTTP(w, r.WithContext(newCtx))
	})
}
