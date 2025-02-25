package app

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/titi0001/Microservices-API-in-Go/src/service"
	"github.com/titi0001/Microservices-API-in-Go/src/domain"
)

func Start() {

	router := mux.NewRouter()

	ch := CustomerHandler{service.NewCustomerService(domain.NewCustomerRepositoryStub())}

	router.HandleFunc("/customers", ch.getAllCustomers).Methods(http.MethodGet)

	err := http.ListenAndServe("localhost:8000", router)
	if err != nil {
		fmt.Println("Error starting server:", err)
	}

}
