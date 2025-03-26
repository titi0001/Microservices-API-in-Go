package repository

import (
	"database/sql"

	"github.com/jmoiron/sqlx"
	"github.com/titi0001/Microservices-API-in-Go/domain"
	"github.com/titi0001/Microservices-API-in-Go/errs"
	"github.com/titi0001/Microservices-API-in-Go/logger"
)

type CustomerRepositoryDb struct {
	client *sqlx.DB
}

func NewCustomerRepositoryDb(dbClient *sqlx.DB) CustomerRepositoryDb {
	return CustomerRepositoryDb{client: dbClient}
}

func (d CustomerRepositoryDb) FindAll(status string) ([]domain.Customer, *errs.AppError) {
	var customers []domain.Customer
	var err error

	if status == "" {
		findAllSQL := "SELECT customer_id, name, date_of_birth, city, zipcode, status FROM customers"
		err = d.client.Select(&customers, findAllSQL)
	} else {
		findAllSQL := "SELECT customer_id, name, date_of_birth, city, zipcode, status FROM customers WHERE status = ?"
		err = d.client.Select(&customers, findAllSQL, status)
	}

	if err != nil {
		logger.Error("Error querying customers", logger.Any("error", err))
		return nil, errs.NewUnexpectedError("Unexpected database error")
	}

	return customers, nil
}

func (d CustomerRepositoryDb) ByID(id string) (*domain.Customer, *errs.AppError) {
	customerSQL := "SELECT customer_id, name, date_of_birth, city, zipcode, status FROM customers WHERE customer_id = ?"
	var c domain.Customer
	err := d.client.Get(&c, customerSQL, id)
	if err != nil {
		if err == sql.ErrNoRows {
			logger.Warn("Customer not found", logger.String("customer_id", id))
			return nil, errs.NewNotFoundError("Customer not found")
		}
		logger.Error("Error fetching customer", logger.Any("error", err))
		return nil, errs.NewUnexpectedError("Unexpected database error")
	}
	return &c, nil
}

func (d CustomerRepositoryDb) Close() {
	if d.client != nil {
		if err := d.client.Close(); err != nil {
			logger.Error("Error closing database connection", logger.Any("error", err))
		}
	}
}
