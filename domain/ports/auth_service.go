package ports

import (
    "github.com/titi0001/Microservices-API-in-Go/api/dto"
    "github.com/titi0001/Microservices-API-in-Go/domain"
    "github.com/titi0001/Microservices-API-in-Go/errs"
)

type AuthService interface {
    RemoteLogin(req dto.LoginRequest) (*dto.LoginResponse, *errs.AppError)
    RemoteIsAuthorized(token, routeName string, vars map[string]string) (bool, *errs.AppError)
    GetSecretKey() []byte
    GetRolePermissions() domain.RolePermissions
}