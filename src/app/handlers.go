package app

import (
	"encoding/json"
	"encoding/xml"
	"net/http"

	"github.com/titi0001/Microservices-API-in-Go/src/service"
)

type Customer struct {
	Name    string `json:"full_name" xml:"name"`
	City    string `json:"city" xml:"city"`
	Zipcode string `json:"zip_code" xml:"zipcode"`
}

type CustomerHandler struct {
	service service.CustomerService
}

func (ch *CustomerHandler) getAllCustomers(w http.ResponseWriter, r *http.Request) {


	customers, error := ch.service.GetAllCustomer()
	if error != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	if r.Header.Get("Content-Type") == "application/xml" {
		w.Header().Add("Content-Type", "application/xml")
		err := xml.NewEncoder(w).Encode(customers)
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
	} else {
		w.Header().Add("Content-Type", "application/json")
		err := json.NewEncoder(w).Encode(customers)
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
	}
}
