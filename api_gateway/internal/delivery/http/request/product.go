package request

type ProductRequest struct {
	ProductCode string   `json:"product_code" validate:"required,alphanum,min=3,max=50"`
	ProductName string   `json:"product_name" validate:"required,min=3,max=150"`
	CategoryID  string   `json:"category_id"  validate:"required,uuid"`
	Price       int64    `json:"price"        validate:"required,gt=0"`
	Stock       int32    `json:"stock"        validate:"gte=0"`
	Description string   `json:"description"  validate:"omitempty,max=500"`
	ImageURLs   []string `json:"image_urls"    validate:"required,dive,min=1"`
}

type FlashSaleRequest struct {
	ProductID          string `json:"product_id" validate:"required,uuid"`
	FlashSalePrice     int64  `json:"flash_sale_price" validate:"required,gt=0"`
	FlashSaleStock     int32  `json:"flash_sale_stock" validate:"gte=0"`
	MaxPurchasePerUser int32  `json:"max_purchase_per_user" validate:"required,gt=0"`
	StartTime          int64  `json:"start_time" validate:"required,gt=0"`
	EndTime            int64  `json:"end_time" validate:"required,gt=0"`
	Version            int32  `json:"version" validate:"omitempty"`
}
