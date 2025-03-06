package service

import "github.com/titi0001/Microservices-API-in-Go/src/domain"

type CustomerService interface {
	GetAllCustomer() ([]domain.Customer, error)
	GetCustomer(string) (*domain.Customer, error)
}

type DefaultCustomerService struct {
	repo domain.CustomerRepository
}

func (s DefaultCustomerService) GetAllCustomer() ([]domain.Customer, error) {
	return s.repo.FindAll()
}

func (s DefaultCustomerService) GetCustomer(id string) (*domain.Customer, error) {
	return s.repo.ById(id)
}

func NewCustomerService(repo domain.CustomerRepository) DefaultCustomerService {
	return DefaultCustomerService{repo}
}
