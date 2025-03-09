package domain

import (
	"github.com/titi0001/Microservices-API-in-Go/src/dto"
	"github.com/titi0001/Microservices-API-in-Go/src/errs"
)

type Customer struct {
	Id          string `db:"customer_id"`
	Name        string
	City        string
	Zipcode     string
	DateOfBirth string `db:"date_of_birth"`
	Status      string
}

func (c Customer) ToDto() dto.CustomerResponse {

	return dto.CustomerResponse{
		Id:          c.Id,
		Name:        c.Name,
		City:        c.City,
		Zipcode:     c.Zipcode,
		DateOfBirth: c.DateOfBirth,
		Status:      c.statusASText(),
	}
}

func (c Customer) statusASText() string {
	statusASText := "active"
	if c.Status == "0" {
		statusASText = "inactive"
	}

	return statusASText
}

type CustomerRepository interface {
	ById(string) (*Customer, *errs.AppError)
	FindAll(status string) ([]Customer, *errs.AppError)
}
