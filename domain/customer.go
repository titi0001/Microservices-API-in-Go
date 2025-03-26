package domain

import (
	"fmt"
    "github.com/titi0001/Microservices-API-in-Go/api/dto"
)

type Customer struct {
    ID          int    `db:"customer_id" json:"customer_id"`
    Name        string `db:"name" json:"name"`
    City        string `db:"city" json:"city"`
    Zipcode     string `db:"zipcode" json:"zipcode"`
    DateOfBirth string `db:"date_of_birth" json:"date_of_birth"`
    Status      int    `db:"status" json:"status"`
}

func (c Customer) ToDto() dto.CustomerResponse {
    return dto.CustomerResponse{
        ID:          fmt.Sprintf("%d", c.ID),
        Name:        c.Name,
        City:        c.City,
        Zipcode:     c.Zipcode,
        DateOfBirth: c.DateOfBirth,
        Status:      c.StatusAsText(), 
    }
}

func (c Customer) StatusAsText() string { 
    if c.Status == 0 {
        return "inactive"
    }
    return "active"
}