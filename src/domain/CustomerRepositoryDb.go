package domain

import (
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"log"
	"os"
	"time"
)

type CustomerRepositoryDb struct {
	client *sql.DB
}

func (d CustomerRepositoryDb) FindAll() ([]Customer, error) {

	rows, err := d.client.Query("SELECT customer_id, name, date_of_birth, city, zipcode, status FROM customers")
	if err != nil {
		panic(err.Error())
	}
	defer rows.Close()

	customers := make([]Customer, 0)
	for rows.Next() {
		var c Customer
		err = rows.Scan(&c.Id, &c.Name, &c.DateOfBirth, &c.City, &c.Zipcode, &c.Status)
		if err != nil {
			panic(err.Error())
		}
		customers = append(customers, c)
	}

	return customers, nil
}

func (d *CustomerRepositoryDb) Close() {
	if d.client != nil {
		d.client.Close()
	}
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

	client, err := sql.Open("mysql", dsn)
	if err != nil {
		panic(err.Error())
	}

	// Testa a conexão com o banco
	if err := client.Ping(); err != nil {
		panic(fmt.Sprintf("Failed to connect to database: %v", err))
	}

	client.SetConnMaxLifetime(time.Minute * 3)
	client.SetMaxOpenConns(10)
	client.SetMaxIdleConns(10)

	return CustomerRepositoryDb{client: client}
}
