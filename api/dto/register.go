package dto

type RegisterRequest struct {
	Username   string `json:"username"`
	Password   string `json:"password"`
	Role       string `json:"role"`
	CustomerID string `json:"customer_id"`
}