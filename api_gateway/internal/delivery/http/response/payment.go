package response

type PaymentResponse struct {
	PaymentID     string `json:"payment_id"`
	OrderID       string `json:"order_id"`
	PaymentStatus string `json:"payment_status"`
	PaymentMethod string `json:"payment_method"`
	PaymentAmount int64  `json:"payment_amount"`
	Message       string `json:"message"`
}
