package domain

import (
	"io"

	"github.com/titi0001/Microservices-API-in-Go/errs"
)

type AuthRepository interface {
	FindUser(username, password string) (*User, *errs.AppError)
	VerifyPermission(role string, customerId interface{}, routeName string, vars map[string]string) bool
}

type AuthService interface {
	RemoteLogin(body io.Reader) ([]byte, *errs.AppError)
	RemoteIsAuthorized(token string, routeName string, vars map[string]string) (bool, *errs.AppError)
	GetRolePermissions() *RolePermissions
	GetSecretKey() []byte
}
