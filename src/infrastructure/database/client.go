// File: src/infrastructure/database/client.go
package database

import (
	"fmt"
	"os"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/go-sql-driver/mysql"
	"github.com/titi0001/Microservices-API-in-Go/src/logger"
)


func GetClient() *sqlx.DB {
	user := os.Getenv("MYSQL_USER")
	password := os.Getenv("MYSQL_PASSWORD")
	dbName := os.Getenv("MYSQL_DATABASE")
	host := "localhost"
	port := "3306"

	if user == "" || password == "" || dbName == "" {
		logger.Fatal("Missing required environment variables",
			logger.String("variables", "MYSQL_USER, MYSQL_PASSWORD, or MYSQL_DATABASE"))
	}
	
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s", user, password, host, port, dbName)

	client, err := sqlx.Open("mysql", dsn)
	if err != nil {
		logger.Error("Error connecting to database", logger.Any("error", err))
		panic(err)
	}

	client.SetConnMaxLifetime(time.Minute * 3)
	client.SetMaxOpenConns(10)
	client.SetMaxIdleConns(10)
	
	return client
}