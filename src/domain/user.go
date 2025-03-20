package domain

import "database/sql"

type User struct {
	Username   string        `db:"username"`
	CustomerId sql.NullInt64 `db:"customer_id"`
	Password   string        `db:"password"`
	Role       string        `db:"role"`
	CreatedOn  string        `db:"created_on"`
}

type UserDetails struct {
	Username string
	UserID   string
	Role     string
	Roles    []string

	CustomClaims map[string]interface{}
}
