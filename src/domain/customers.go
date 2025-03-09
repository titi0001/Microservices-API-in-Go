package domain

import "github.com/titi0001/Microservices-API-in-Go/src/errs"

type Customer struct {
	Id          string
	Name        string
	City        string
	Zipcode     string
	DateOfBirth string
	Status      string
}

type CustomerRepository interface {
	
	ById(string) (*Customer, *errs.AppError)
	FindAll(status string) ([]Customer, *errs.AppError)
}
