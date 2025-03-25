package dto

type User struct {
	Username   string `json:"username"`
	Role       string `json:"role"`
	CustomerID string `json:"customer_id"`
	CreatedOn  string `json:"created_on"`
}