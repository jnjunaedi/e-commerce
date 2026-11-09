package response

import "time"

type OrderResponse struct {
	ID             string      `json:"order_id"`
	UserID         string      `json:"user_id"`
	TotalPrice     int64       `json:"total_price"`
	OrderStatus    string      `json:"order_status"`
	IdempotencyKey string      `json:"idempotency_key"`
	Items          []OrderItem `json:"items"`
	CreatedAt      time.Time   `json:"created_at"`
}

type OrderItem struct {
	ID          string `json:"item_id"`
	OrderID     string `json:"order_id"`
	ProductID   string `json:"product_id"`
	FlashSaleID string `json:"flash_sale_id"`
	Quantity    int32  `json:"quantity"`
	UnitPrice   int64  `json:"unit_price"`
}
