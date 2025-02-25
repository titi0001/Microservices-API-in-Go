package app

import (
	"fmt"
	"net/http"
)

func Start() {

	http.HandleFunc("/greet", greet)
	http.HandleFunc("/customers", getAllCustomers)

	err := http.ListenAndServe("localhost:8000", nil)
	if err != nil {
		fmt.Println("Error starting server:", err)
	}

}
