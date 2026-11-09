package domain

import (
	"context"
	"time"
)

type Notification struct {
	ID               string    `json:"id"`
	UserID           string    `json:"user_id"`
	Email            string    `json:"email"`
	OrderID          string    `json:"order_id"`
	NotificationType string    `json:"notification_type"`
	Title            string    `json:"title"`
	Message          string    `json:"message"`
	Status           string    `json:"status"`
	CreatedAt        time.Time `json:"created_at"`
}

type Repository interface {
	Create(ctx context.Context, n Notification) error
	UpdateStatus(ctx context.Context, id string, status string) error
}

type Usecase interface {
	SendAndSaveNotification(ctx context.Context, n Notification) error
}
