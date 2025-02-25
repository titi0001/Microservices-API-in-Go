package app

import (
	"fmt"
	"github.com/gorilla/mux"
	"net/http"
)

func Start() {

	router := mux.NewRouter()

	router.HandleFunc("/greet", greet).Methods(http.MethodGet)
	router.HandleFunc("/customers", getAllCustomers).Methods(http.MethodGet)
	router.HandleFunc("/customers", createCustomer).Methods(http.MethodPost)
	router.HandleFunc("/customers/{custormer_id:[0-9]+}", getCustomers).Methods(http.MethodGet)

	err := http.ListenAndServe("localhost:8000", router)
	if err != nil {
		fmt.Println("Error starting server:", err)
	}

}

func getCustomers(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	fmt.Fprint(w,  vars["customer_id"])
}

func createCustomer(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "Post request received")
}