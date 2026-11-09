package request

type PaymentRequest struct {
	OrderID       string `json:"order_id" validate:"required"`
	PaymentMethod string `json:"payment_method" validate:"required"`
	PaymentAmount int64  `json:"payment_amount" validate:"required,gt=0"`
}

type MidtransWebhookRequest struct {
	OrderID           string `json:"order_id" validate:"required"`
	TransactionStatus string `json:"transaction_status" validate:"required"`
	FraudStatus       string `json:"fraud_status" validate:"required"`
	SignatureKey      string `json:"signature_key" validate:"required"`
	GrossAmount       string `json:"gross_amount" validate:"required"`
}
