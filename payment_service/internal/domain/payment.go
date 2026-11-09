package domain

import (
	"context"
	"time"
)

type Payment struct {
	ID            string
	OrderID       string
	PaymentMethod string
	PaymentAmount int64
	PaymentStatus string
	PaidAt        int64
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

type TxManager interface {
	WithTransaction(ctx context.Context, fn func(ctx context.Context) error) error
}

type Repository interface {
	Create(ctx context.Context, p Payment) error
	GetByOrderID(ctx context.Context, orderID string) (Payment, error)
	UpdateStatus(ctx context.Context, orderID, status string, paidAt int64) error
	CreateOutbox(ctx context.Context, id, eventType string, payload []byte) error
}

type Usecase interface {
	Process(ctx context.Context, p Payment) (Payment, error)
	HandleWebhook(ctx context.Context, payload map[string]interface{}) error
}
