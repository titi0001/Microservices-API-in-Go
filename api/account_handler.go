package api

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/titi0001/Microservices-API-in-Go/api/dto"
	"github.com/titi0001/Microservices-API-in-Go/domain/ports"
	"github.com/titi0001/Microservices-API-in-Go/infrastructure/utils"
	"github.com/titi0001/Microservices-API-in-Go/logger"
)

type AccountHandler struct {
	service ports.AccountService
}

func NewAccountHandler(service ports.AccountService) *AccountHandler {
	return &AccountHandler{service: service}
}

func (h *AccountHandler) NewAccount(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		logger.Warn("Invalid method for NewAccount", logger.String("method", r.Method))
		utils.WriteResponse(w, http.StatusMethodNotAllowed, map[string]string{"error": "Method not allowed"})
		return
	}

	vars := mux.Vars(r)
	customerID := vars["customer_id"]

	var request dto.NewAccountRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		logger.Warn("Failed to decode NewAccount request", logger.Any("error", err))
		utils.WriteResponse(w, http.StatusBadRequest, map[string]string{"error": "Invalid request payload"})
		return
	}

	request.CustomerID = customerID
	if err := request.Validate(); err != nil {
		logger.Warn("Validation failed for NewAccount", logger.Any("error", err))
		utils.WriteResponse(w, err.Code, map[string]string{"error": err.AsMessage()})
		return
	}

	response, appError := h.service.NewAccount(request)
	if appError != nil {
		logger.Error("Error creating new account", logger.Any("error", appError))
		utils.WriteResponse(w, appError.Code, map[string]string{"error": appError.AsMessage()})
		return
	}
	utils.WriteResponse(w, http.StatusCreated, response)
}

func (h *AccountHandler) MakeTransaction(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		logger.Warn("Invalid method for MakeTransaction", logger.String("method", r.Method))
		utils.WriteResponse(w, http.StatusMethodNotAllowed, map[string]string{"error": "Method not allowed"})
		return
	}

	vars := mux.Vars(r)
	accountID := vars["account_id"]
	customerID := vars["customer_id"]

	var request dto.TransactionRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		logger.Warn("Failed to decode Transaction request", logger.Any("error", err))
		utils.WriteResponse(w, http.StatusBadRequest, map[string]string{"error": "Invalid request payload"})
		return
	}

	request.AccountID = accountID
	request.CustomerID = customerID
	if err := request.Validate(); err != nil {
		logger.Warn("Validation failed for Transaction", logger.Any("error", err))
		utils.WriteResponse(w, err.Code, map[string]string{"error": err.AsMessage()})
		return
	}

	response, appError := h.service.MakeTransaction(request)
	if appError != nil {
		logger.Error("Error processing transaction", logger.Any("error", appError))
		utils.WriteResponse(w, appError.Code, map[string]string{"error": appError.AsMessage()})
		return
	}
	utils.WriteResponse(w, http.StatusOK, response)
}