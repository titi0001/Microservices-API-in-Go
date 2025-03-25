package api

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/titi0001/Microservices-API-in-Go/domain/ports"
	"github.com/titi0001/Microservices-API-in-Go/infrastructure/utils"
)

type CustomerHandler struct {
	service ports.CustomerService
}

func NewCustomerHandler(service ports.CustomerService) *CustomerHandler {
	return &CustomerHandler{service: service}
}

func (ch *CustomerHandler) GetAllCustomers(w http.ResponseWriter, r *http.Request) {
	status := r.URL.Query().Get("status")

	customers, err := ch.service.GetAllCustomer(status)
	if err != nil {
		utils.WriteResponse(w, err.Code, err.AsMessage())
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
		return
	}
	utils.WriteResponse(w, http.StatusOK, customer)
}