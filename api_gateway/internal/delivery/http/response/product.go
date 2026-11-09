package response

import "time"

type ProductResponse struct {
	ID          string    `json:"product_id"`
	ProductCode string    `json:"product_code"`
	ProductName string    `json:"product_name"`
	CategoryID  string    `json:"category_id"`
	Price       int64     `json:"price"`
	Stock       int32     `json:"stock"`
	Description string    `json:"description"`
	ImageURLs   []string  `json:"image_urls"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type FlashSaleResponse struct {
	ProductID          string `json:"product_id"`
	FlashSalePrice     int64  `json:"flash_sale_price"`
	FlashSaleStock     int32  `json:"flash_sale_stock"`
	MaxPurchasePerUser int32  `json:"max_purchase_per_user"`
	StartTime          int64  `json:"start_time"`
	EndTime            int64  `json:"end_time"`
	Version            int32  `json:"version"`
}
