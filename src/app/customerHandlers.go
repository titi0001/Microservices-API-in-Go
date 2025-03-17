package app

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/titi0001/Microservices-API-in-Go/src/infrastructure/utils"
	"github.com/titi0001/Microservices-API-in-Go/src/service"
)

type CustomerHandler struct {
	service service.CustomerService
}

func (ch *CustomerHandler) getAllCustomers(w http.ResponseWriter, r *http.Request) {
	status := r.URL.Query().Get("status")

	customers, err := ch.service.GetAllCustomer(status)
	if err != nil {
		http.Error(w, err.Message, err.Code)
		return
	}

	utils.WriteResponse(w, http.StatusOK, customers)  
}

func (ch *CustomerHandler) GetCustomer(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["customer_id"]

	customer, err := ch.service.GetCustomer(id)
	if err != nil {
		utils.WriteResponse(w, err.Code, err.AsMessage())  
	} else {
		utils.WriteResponse(w, http.StatusOK, customer)  
	}
}