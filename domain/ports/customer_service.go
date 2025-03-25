package ports

import (
	"github.com/titi0001/Microservices-API-in-Go/api/dto"
	"github.com/titi0001/Microservices-API-in-Go/errs"
)

type CustomerService interface {
	GetCustomer(id string) (*dto.CustomerResponse, *errs.AppError)
	GetAllCustomer(status string) ([]dto.CustomerResponse, *errs.AppError)
}