package ports

import (
	"github.com/titi0001/Microservices-API-in-Go/domain"
	"github.com/titi0001/Microservices-API-in-Go/errs"
)

type CustomerRepository interface {
	ByID(id string) (*domain.Customer, *errs.AppError)
	FindAll(status string) ([]domain.Customer, *errs.AppError)
}