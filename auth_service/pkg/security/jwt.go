package security

import (
	"e-commerce/pkg/utils/apperror"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

var (
	AccessTokenSecret  = []byte(os.Getenv("JWT_ACCESS_SECRET"))
	RefreshTokenSecret = []byte(os.Getenv("JWT_REFRESH_SECRET"))
)

type TokenProvider interface {
	GenerateAccessToken(userID, email, username string) (string, error)
	GenerateRefreshAccessToken(userID string) (string, error)
	VerifyAccessToken(tokenString string) (*TokenClaims, error)
	VerifyRefreshToken(tokenString string) (string, error)
}

type TokenClaims struct {
	UserID   string `json:"user_id"`
	Email    string `json:"email"`
	Username string `json:"username"`
	jwt.RegisteredClaims
}

func NewJwtProvider() TokenProvider {
	return &TokenClaims{}
}

func (j *TokenClaims) GenerateAccessToken(userID, email, username string) (string, error) {
	tokenID, err := uuid.NewV7()
	if err != nil {
		return "", err
	}

	claims := TokenClaims{
		UserID:   userID,
		Email:    email,
		Username: username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(15 * time.Minute)),
			Issuer:    "jjn-ecommerce",
			Subject:   userID,
			ID:        tokenID.String(),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(AccessTokenSecret)
}

func (j *TokenClaims) GenerateRefreshAccessToken(userID string) (string, error) {
	claims := jwt.RegisteredClaims{
		Subject:   userID,
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(7 * 24 * time.Hour)),
		IssuedAt:  jwt.NewNumericDate(time.Now()),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(RefreshTokenSecret)
}

func (j *TokenClaims) VerifyAccessToken(tokenString string) (*TokenClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &TokenClaims{}, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, apperror.New(apperror.ErrUnauthorized, "metode enkripsi token tidak sesuai")
		}
		return AccessTokenSecret, nil
	})

	if err != nil {
		return &TokenClaims{}, err
	}

	if claims, ok := token.Claims.(*TokenClaims); ok && token.Valid {
		return &TokenClaims{
			UserID:   claims.UserID,
			Email:    claims.Email,
			Username: claims.Username,
		}, nil
	}

	return &TokenClaims{}, apperror.New(apperror.ErrUnauthorized, "token tidak valid")
}

func (j *TokenClaims) VerifyRefreshToken(tokenString string) (string, error) {
	token, err := jwt.Parse(tokenString, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, apperror.New(apperror.ErrUnauthorized, "metode enkripsi token tidak sesuai")
		}
		return RefreshTokenSecret, nil
	})

	if err != nil {
		return "", err
	}

	if token.Valid {
		userID, err := token.Claims.GetSubject()
		if err != nil {
			return "", err
		}
		return userID, nil
	}

	return "", apperror.New(apperror.ErrUnauthorized, "refresh token tidak valid")
}
