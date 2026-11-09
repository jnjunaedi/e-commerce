package request

type UserProfile struct {
	UserID      string `json:"user_id" validate:"required,uuid"`
	Username    string `json:"username" validate:"required,min=3,max=15"`
	FullName    string `json:"full_name" validate:"required,min=3,max=100"`
	PhoneNumber string `json:"phone_number" validate:"required,min=10,max=12"`
	AvatarURL   string `json:"avatar_url" validate:"omitempty"`
}

type UserAddress struct {
	UserID        string `json:"user_id" validate:"required,uuid"`
	RecipientName string `json:"recipient_name" validate:"required,min=3,max=100"`
	PhoneNumber   string `json:"phone_number" validate:"required,min=10,max=12"`
	FullAddress   string `json:"full_address" validate:"required,min=3"`
	City          string `json:"city" validate:"required,min=3"`
	PostalCode    string `json:"postal_code" validate:"required,max=5"`
	IsDefault     bool   `json:"is_default" validate:"required"`
}
