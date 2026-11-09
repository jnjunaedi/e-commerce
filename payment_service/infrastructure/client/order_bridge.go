package client

import (
	"context"
	orderpb "e-commerce/pb/order"
	"time"
)

type OrderResult struct {
	ID             string
	UserID         string
	TotalPrice     int64
	OrderStatus    string
	IdempotencyKey string
	Items          []OrderItemResult
	CreatedAt      time.Time
}

type OrderItemResult struct {
	ID          string
	OrderID     string
	ProductID   string
	FlashSaleID string
	Quantity    int32
	UnitPrice   int64
	Subtotal    int64
}

type OrderGrpcClient interface {
	GetOrderDetails(ctx context.Context, orderID string) (OrderResult, error)
	UpdateOrderStatus(ctx context.Context, orderID, status string) error
}

type OrderClientBridgeImpl struct {
	orderGRPC orderpb.OrdersServiceClient
}

func NewOrderClient(orderGRPC orderpb.OrdersServiceClient) OrderGrpcClient {
	return &OrderClientBridgeImpl{orderGRPC: orderGRPC}
}

func (c *OrderClientBridgeImpl) GetOrderDetails(ctx context.Context, orderID string) (OrderResult, error) {
	res, err := c.orderGRPC.GetOrder(ctx, &orderpb.GetOrderRequest{OrderId: orderID})
	if err != nil {
		return OrderResult{}, err
	}

	items := make([]OrderItemResult, 0, len(res.GetItems()))
	for _, item := range res.GetItems() {
		items = append(items, OrderItemResult{
			ID:          item.GetItemId(),
			ProductID:   item.GetProductId(),
			FlashSaleID: item.GetFlashSaleId(),
			Quantity:    item.GetQuantity(),
			UnitPrice:   item.GetUnitPrice(),
		})
	}

	return OrderResult{
		ID:             res.GetOrderId(),
		UserID:         res.GetUserId(),
		TotalPrice:     res.GetTotalPrice(),
		OrderStatus:    res.GetOrderStatus(),
		IdempotencyKey: res.GetIdempotencyKey(),
		Items:          items,
		CreatedAt:      res.GetCreatedAt().AsTime(),
	}, nil
}

func (c *OrderClientBridgeImpl) UpdateOrderStatus(ctx context.Context, orderID string, status string) error {
	_, err := c.orderGRPC.UpdateOrderStatus(ctx, &orderpb.UpdateOrderStatusRequest{OrderId: orderID, Status: status})
	if err != nil {
		return err
	}

	return nil
}
