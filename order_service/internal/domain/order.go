package domain

import (
	"context"
	"time"
)

type Order struct {
	ID             string
	UserID         string
	TotalPrice     int64
	OrderStatus    string
	IdempotencyKey string
	Items          []OrderItem
	CreatedAt      time.Time
}

type OrderItem struct {
	ID          string
	OrderID     string
	ProductID   string
	FlashSaleID *string
	Quantity    int32
	UnitPrice   int64
	Subtotal    int64
}

type TxManager interface {
	WithTransaction(ctx context.Context, fn func(ctx context.Context) error) error
}

type Repository interface {
	Create(ctx context.Context, order Order) error
	GetByID(ctx context.Context, id string) (Order, error)
	CheckByIdempotencyKey(ctx context.Context, key string) (bool, error)
	UpdateStatus(ctx context.Context, orderID, status string) error
	CreateOutbox(ctx context.Context, id, eventType string, payload []byte) error
	DeductLocalFlashSaleStock(ctx context.Context, flashSaleID string, qty int32) (int64, error)
}

type Usecase interface {
	CreateCheckout(ctx context.Context, order Order) (Order, error)
	GetOrderDetails(ctx context.Context, orderID string) (Order, error)
	UpdateOrderStatus(ctx context.Context, orderID string, status string) error
}
