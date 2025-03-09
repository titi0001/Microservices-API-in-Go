package domain

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	"github.com/titi0001/Microservices-API-in-Go/src/errs"
	"github.com/titi0001/Microservices-API-in-Go/src/logger"
)

type CustomerRepositoryDb struct {
	client *sqlx.DB
}

func (d CustomerRepositoryDb) FindAll(status string) ([]Customer, *errs.AppError) {
	customers := make([]Customer, 0)

	if status != "" {
		findAllSql := "SELECT customer_id, name, date_of_birth, city, zipcode, status FROM customers WHERE status = ?"
		err := d.client.Select(&customers, findAllSql)
		if err != nil {
			logger.Error("Error while scanning customer" + err.Error())
		}
	} else {
		findAllSql := "SELECT customer_id, name, date_of_birth, city, zipcode, status FROM customers WHERE status = ?"
		err := d.client.Select(&customers, findAllSql, status)
		if err != nil {
			logger.Error("Error while scanning customer" + err.Error())
		}
	}
	return customers, nil
}

func (d CustomerRepositoryDb) ById(id string) (*Customer, *errs.AppError) {
	customerSql := "SELECT customer_id, name, date_of_birth, city, zipcode, status FROM customers WHERE customer_id = ?"

	var c Customer
	err := d.client.Get(&c, customerSql, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errs.NewNotFoundError("Customer not found")
		}
		log.Println("Error getting customer:", err.Error())
		return nil, errs.NewUnexpectedError("Unexpected database error")
	}
	return &c, nil
}

func NewCustomerRepositoryDb() CustomerRepositoryDb {
	user := os.Getenv("MYSQL_USER")
	password := os.Getenv("MYSQL_PASSWORD")
	dbName := os.Getenv("MYSQL_DATABASE")
	host := "localhost"
	port := "3306"

	if user == "" || password == "" || dbName == "" {
		log.Fatal("Missing required environment variables: MYSQL_USER, MYSQL_PASSWORD, or MYSQL_DATABASE")
	}
	// string de conexão .env
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s", user, password, host, port, dbName)

	client, err := sqlx.Open("mysql", dsn)
	if err != nil {
		logger.Error("Error connecting to database" + err.Error())
	}

	client.SetConnMaxLifetime(time.Minute * 3)
	client.SetMaxOpenConns(10)
	client.SetMaxIdleConns(10)

	return CustomerRepositoryDb{client}
}

func (d *CustomerRepositoryDb) Close() {
	if d.client != nil {
		d.client.Close()
	}
}
