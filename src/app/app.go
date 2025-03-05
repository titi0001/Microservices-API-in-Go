package app

import (
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"log"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	"github.com/titi0001/Microservices-API-in-Go/src/domain"
	"github.com/titi0001/Microservices-API-in-Go/src/service"
)

func Start() {

	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}


	router := mux.NewRouter()
	repo := domain.NewCustomerRepositoryDb()

	ch := CustomerHandler{service.NewCustomerService(repo)}

	router.HandleFunc("/customers", ch.getAllCustomers).Methods(http.MethodGet)

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
	repo.Close()

}
