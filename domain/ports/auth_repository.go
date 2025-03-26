package ports

import (
	"github.com/titi0001/Microservices-API-in-Go/domain"
	"github.com/titi0001/Microservices-API-in-Go/errs"
)

type AuthRepository interface {
	FindUser(username, password string) (*domain.User, *errs.AppError)
	VerifyPermission(role string, customerId string, routeName string, vars map[string]string) bool
	SaveUser(user domain.User) (*domain.User, *errs.AppError)
	SaveRefreshToken(refreshToken string) *errs.AppError
	VerifyRefreshToken(refreshToken string) (bool, *errs.AppError)
}
