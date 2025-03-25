package domain

import (
	"github.com/titi0001/Microservices-API-in-Go/api/dto"
)

type Customer struct {
	ID          string `json:"customer_id"`
	Name        string `json:"name"`
	City        string `json:"city"`
	Zipcode     string `json:"zipcode"`
	DateOfBirth string `json:"date_of_birth"`
	Status      string `json:"status"`
}

func (c Customer) ToDto() dto.CustomerResponse {
	return dto.CustomerResponse{
		ID:          c.ID,
		Name:        c.Name,
		City:        c.City,
		Zipcode:     c.Zipcode,
		DateOfBirth: c.DateOfBirth,
		Status:      c.statusAsText(),
	}
}

func (c Customer) statusAsText() string {
	if c.Status == "0" {
		return "inactive"
	}
	return "active"
}