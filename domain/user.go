package domain

import "time"

type User struct {
	ID         string    `db:"id" json:"id,omitempty"`
	Username   string    `db:"username" json:"username"`
	Password   string    `db:"password" json:"password"`
	Role       string    `db:"role" json:"role"`
	CustomerID *string   `db:"customer_id" json:"customer_id,omitempty"`
	CreatedOn  time.Time `db:"created_on" json:"created_on"`
}
