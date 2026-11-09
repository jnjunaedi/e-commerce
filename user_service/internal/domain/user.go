package domain

import (
	"context"
	"time"
)

type Profile struct {
	ID          string
	UserID      string
	Username    string
	FullName    string
	PhoneNumber string
	AvatarURL   string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type Address struct {
	ID            string
	UserID        string
	RecipientName string
	Phone         string
	FullAddress   string
	City          string
	PostalCode    string
	IsDefault     bool
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

type Repository interface {
	CreateProfile(ctx context.Context, profile Profile) error
	UpdateProfile(ctx context.Context, profile Profile) error
	CreateAddress(ctx context.Context, address Address) error
	UpdateAddress(ctx context.Context, address Address) error
	GetAddressByUserID(ctx context.Context, userID string) ([]Address, error)
}

type Usecase interface {
	RegisterProfile(ctx context.Context, profile Profile) error
	UpdateProfile(ctx context.Context, profile Profile) error
	AddAddress(ctx context.Context, address Address) error
	UpdateAddress(ctx context.Context, address Address) error
	FindAddressByUserID(ctx context.Context, userID string) ([]Address, error)
}
