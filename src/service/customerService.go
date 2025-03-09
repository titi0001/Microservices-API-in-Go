package service

import (
	"github.com/titi0001/Microservices-API-in-Go/src/domain"
	"github.com/titi0001/Microservices-API-in-Go/src/errs"
)

func NewCustomerService(repo domain.CustomerRepository) CustomerService {
    return DefaultCustomerService{repo: repo}
}


type CustomerService interface {
	GetCustomer(string) (*domain.Customer, *errs.AppError)
	GetAllCustomer(status string) ([]domain.Customer, *errs.AppError)
}

type DefaultCustomerService struct {
	repo domain.CustomerRepository
}

func (s DefaultCustomerService) GetAllCustomer(status string ) ([]domain.Customer, *errs.AppError) {
	if status == "active" {
		status = "1"
	} else if status == "inactive" {
		status = "0"
	} else {	
		status = ""
	}
	return s.repo.FindAll(status)
}


func (s DefaultCustomerService) GetCustomer(id string) (*domain.Customer, *errs.AppError) {
	return s.repo.ById(id)
}

