package domain

import (
	"context"
	"time"
)

type Product struct {
	ID          string
	ProductCode string
	ProductName string
	CategoryID  string
	Price       int64
	Stock       int32
	Description string
	ImageURLs   []string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type ProductFlashSale struct {
	ID                 string
	ProductID          string
	FlashSalePrice     int64
	FlashSaleStock     int32
	MaxPurchasePerUser int32
	StartTime          int64
	EndTime            int64
	Version            int32
}

type Repository interface {
	Create(ctx context.Context, p Product) error
	GetByID(ctx context.Context, id string) (Product, error)
	UpdateRegularStock(ctx context.Context, id string, quantity int32) error
	CreateFlashSale(ctx context.Context, fs ProductFlashSale) error
	GetFlashSaleByID(ctx context.Context, id string) (ProductFlashSale, error)
	UpdateFlashSaleStock(ctx context.Context, id string, quantity int32, currentVersion int32) error
	SubscribeStockChanges(ctx context.Context, flashSaleID string) (<-chan int32, error)
	BroadcastStockUpdate(flashSaleID string, latestStock int32)
}

type Usecase interface {
	RegisterProduct(ctx context.Context, p Product) (Product, error)
	FindProduct(ctx context.Context, id string) (Product, error)
	DeductRegularStock(ctx context.Context, productID string, qty int32) error
	StartFlashSaleEvent(ctx context.Context, fs ProductFlashSale) (ProductFlashSale, error)
	FindFlashSale(ctx context.Context, id string) (ProductFlashSale, error)
	DeductFlashSaleStock(ctx context.Context, productID string, flashSaleID string, qty int32) (bool, string, error)
	StreamStock(ctx context.Context, flashSaleID string) (<-chan int32, error)
}
