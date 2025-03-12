package app

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gorilla/mux"
	"github.com/jmoiron/sqlx"
	"github.com/joho/godotenv"
	"github.com/titi0001/Microservices-API-in-Go/src/domain"
	"github.com/titi0001/Microservices-API-in-Go/src/logger"
	"github.com/titi0001/Microservices-API-in-Go/src/service"
)

func Start() {

	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	router := mux.NewRouter()
	dbClient := getDbClient()
	CustomerRepositoryDb := domain.NewCustomerRepositoryDb(dbClient)
	accountRepositoryDB := domain.NewAccountRepositoryDb(dbClient)
	customerService := service.NewCustomerService(CustomerRepositoryDb)

	ch := CustomerHandler{service: customerService}
	ah := AccountHandler{service: service.NewAccountService(accountRepositoryDB)}

	router.HandleFunc("/customers", ch.getAllCustomers).Methods(http.MethodGet)
	router.HandleFunc("/customers/{customer_id:[0-9]+}", ch.GetCustomer).Methods(http.MethodGet)
	router.HandleFunc("/customers/{customer_id:[0-9]+}/account", ah.NewAccount).Methods(http.MethodPost)
	router.
		HandleFunc("/customers/{customer_id:[0-9]+}/account/{account_id:[0-9]+}", ah.MakeTransaction).Methods(http.MethodPost)

	server := &http.Server{
		Addr:    "localhost:8000",
		Handler: router,
	}

	go func() {
		fmt.Println("Server started on localhost:8000")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			fmt.Println("Error starting server:", err)
		}
	}()

	// Aguardar sinal para fechar o servidor
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	<-sigChan

	// Fechar o banco de dados
	fmt.Println("Shutting down server...")
	CustomerRepositoryDb.Close()

}

func getDbClient() *sqlx.DB {
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
	return client
}
