package domain

import (
	"context"
	"time"
)

type Auth struct {
	ID           string
	Email        string
	Username     string
	PasswordHash string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type RegisterRequest struct {
	Email    string
	Username string
	Password string
}

type LoginRequest struct {
	Email    string
	Password string
}

type TokenResponse struct {
	AccessToken  string
	RefreshToken string
}

type TxManager interface {
	WithTransaction(ctx context.Context, fn func(ctx context.Context) error) error
}

type Repository interface {
	Create(ctx context.Context, auth Auth) error
	FindByEmail(ctx context.Context, email string) (Auth, error)
	ExistsByEmailOrUsername(ctx context.Context, email, username string) (bool, error)
	CreateOutbox(ctx context.Context, id, eventType string, payload []byte) error
}

type Usecase interface {
	Register(ctx context.Context, req RegisterRequest) (string, error)
	Login(ctx context.Context, req LoginRequest) (TokenResponse, error)
	ValidateToken(ctx context.Context, token string) (Auth, error)
}
