package models

import "time"

const (
	RoleAdmin  = "admin"
	RoleClient = "client"
)

const (
	StatusPending    = "en attente"
	StatusPaid       = "payé"
	StatusDelivering = "en cours de livraison"
	StatusDelivered  = "livré"
	StatusCancelled  = "annulé"
)

type User struct {
	ID               int       `json:"id"`
	Email            string    `json:"email"`
	PasswordHash     string    `json:"-"`
	Password         string    `json:"password,omitempty"`
	Role             string    `json:"role"`
	IsConfirmed      bool      `json:"is_confirmed"`
	ConfirmationCode string    `json:"confirmation_code,omitempty"`
	ResetCode        string    `json:"reset_code,omitempty"`
	CreatedAt        time.Time `json:"created_at"`
}

type RegisterRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type ConfirmRequest struct {
	Email string `json:"email"`
	Code  string `json:"code"`
}

type ForgotPasswordRequest struct {
	Email string `json:"email"`
}

type ResetPasswordRequest struct {
	Email       string `json:"email"`
	Code        string `json:"code,omitempty"`
	NewPassword string `json:"new_password"`
}

type AuthResponse struct {
	Token            string `json:"token,omitempty"`
	Message          string `json:"message,omitempty"`
	ConfirmationCode string `json:"confirmation_code,omitempty"`
	ResetCode        string `json:"reset_code,omitempty"`
	User             *User  `json:"user,omitempty"`
}

type Product struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Category    string    `json:"category"`
	PriceHT     float64   `json:"price_ht"`
	PriceTTC    float64   `json:"price_ttc"`
	Stock       int       `json:"stock"`
	CreatedAt   time.Time `json:"created_at"`
}

type CartItem struct {
	ID        int     `json:"id,omitempty"`
	ProductID string  `json:"product_id"`
	Product   Product `json:"product,omitempty"`
	Quantity  int     `json:"quantity"`
}

type Cart struct {
	ID        string     `json:"id"`
	UserID    int        `json:"user_id"`
	Items     []CartItem `json:"items"`
	TotalTTC  float64    `json:"total_ttc"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
}

type OrderItem struct {
	ID           int     `json:"id,omitempty"`
	OrderID      string  `json:"order_id"`
	ProductID    string  `json:"product_id"`
	ProductName  string  `json:"product_name,omitempty"`
	Quantity     int     `json:"quantity"`
	PriceUnitTTC float64 `json:"price_unit_ttc"`
}

type Order struct {
	ID           string      `json:"id"`
	UserID       int         `json:"user_id"`
	UserEmail    string      `json:"user_email,omitempty"`
	Status       string      `json:"status"`
	CancelReason string      `json:"cancel_reason,omitempty"`
	TotalTTC     float64     `json:"total_ttc"`
	Items        []OrderItem `json:"items,omitempty"`
	CreatedAt    time.Time   `json:"created_at"`
}

type UpdateOrderStatusRequest struct {
	Status       string `json:"status"`
	CancelReason string `json:"cancel_reason,omitempty"`
}

type CreateManualOrderRequest struct {
	UserID int `json:"user_id"`
	Items  []struct {
		ProductID string `json:"product_id"`
		Quantity  int    `json:"quantity"`
	} `json:"items"`
}
