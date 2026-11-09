package request

type OrderItems struct {
	ProductID   string `json:"product_id" validate:"required,uuid"`
	FlashSaleID string `json:"flash_sale_id" validate:"omitempty,uuid"`
	Quantity    int32  `json:"quantity" validate:"required,gt=0"`
}

type OrderRequest struct {
	UserID         string       `json:"user_id" validate:"required,uuid"`
	IdempotencyKey string       `json:"idempotency_key" validate:"required,min=5"`
	Items          []OrderItems `json:"items" validate:"required,dive"`
}
