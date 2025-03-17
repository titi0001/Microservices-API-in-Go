package utils

import (
	"encoding/json"
	"net/http"
	
	"github.com/titi0001/Microservices-API-in-Go/src/logger"
)

func WriteResponse(w http.ResponseWriter, code int, data any) {
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(code)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		logger.Error("Failed to encode response", logger.Any("error", err))
	}
}