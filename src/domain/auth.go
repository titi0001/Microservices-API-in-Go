package domain

import (
	"database/sql"
	"io"

	"github.com/titi0001/Microservices-API-in-Go/src/errs"
)

type User struct {
	Username   string        `db:"username"`
	Password   string        `db:"password"`
	Role       string        `db:"role"`
	CustomerId sql.NullInt64 `db:"customer_id"`
	CreatedOn  string        `db:"created_on"`
}

type AuthRepository interface {
	FindUser(username, password string) (*User, *errs.AppError)
	VerifyPermission(role string, customerId interface{}, routeName string, vars map[string]string) bool
}

type AuthService interface {
	RemoteLogin(body io.Reader) ([]byte, *errs.AppError)
	RemoteIsAuthorized(token string, routeName string, vars map[string]string) (bool, *errs.AppError)
}
