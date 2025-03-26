package service

import (
	"fmt"
	"github.com/titi0001/Microservices-API-in-Go/api/dto"
	"github.com/titi0001/Microservices-API-in-Go/domain/ports"
	"github.com/titi0001/Microservices-API-in-Go/errs"
)

type DefaultCustomerService struct {
	repo ports.CustomerRepository
}

func NewCustomerService(repo ports.CustomerRepository) ports.CustomerService {
	return &DefaultCustomerService{repo: repo}
}

func (s DefaultCustomerService) GetAllCustomer(status string) ([]dto.CustomerResponse, *errs.AppError) {
	switch status {
	case "active":
		status = "1"
	case "inactive":
		status = "0"
	default:
		status = ""
	}

	customers, err := s.repo.FindAll(status)
	if err != nil {
		return nil, err
	}

	response := make([]dto.CustomerResponse, 0, len(customers))
	for _, c := range customers {
		response = append(response, c.ToDto())
	}
	return response, nil
}

func (s DefaultCustomerService) GetCustomer(id string) (*dto.CustomerResponse, *errs.AppError) {
	c, err := s.repo.ByID(id)
	if err != nil {
		return nil, err
	}
	return &dto.CustomerResponse{
		ID:          fmt.Sprintf("%d", c.ID),
		Name:        c.Name,
		City:        c.City,
		Zipcode:     c.Zipcode,
		DateOfBirth: c.DateOfBirth,
		Status:      c.StatusAsText(),
	}, nil
}
