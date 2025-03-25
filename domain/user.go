package domain

type User struct {
	ID          string `json:"id"`
	Username    string `json:"username"`
	Password    string `json:"password"`
	Role        string `json:"role"`
	CustomerID  string `json:"customer_id"`
	CreatedOn   string `json:"created_on"`
}