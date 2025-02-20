package main

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type Customer struct {
	Name    string
	City    string
	Zipcode string
}

func main() {
	http.HandleFunc("/greet", greet)
	http.HandleFunc("/customers", getAllCustomers)

	err := http.ListenAndServe("localhost:8000", nil)
	if err != nil {
		fmt.Println("Error starting server:", err)
	}
}

func greet(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "Hello World")
}

func getAllCustomers(w http.ResponseWriter, r *http.Request) {
	customers := []Customer{
		{Name: "John Doe", City: "New York", Zipcode: "10001"},
		{Name: "Jane Doe", City: "New York", Zipcode: "10001"},
	}


    w.Header().Add("Content-Type", "application/json")
	json.NewEncoder(w).Encode(customers)
}
