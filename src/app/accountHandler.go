package app

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/titi0001/Microservices-API-in-Go/src/dto"
	"github.com/titi0001/Microservices-API-in-Go/src/infrastructure/utils"
	"github.com/titi0001/Microservices-API-in-Go/src/service"
)

type AccountHandler struct {
	service service.AccountService
}

func (h AccountHandler) NewAccount(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	customerId := vars["customer_id"]
	var request dto.NewAccountRequest
	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		utils.WriteResponse(w, http.StatusBadRequest, err.Error())
	} else {
		request.CustomerId = customerId
		account, appError := h.service.NewAccount(request)
		if appError != nil {
			utils.WriteResponse(w, appError.Code, appError.Message)
		} else {
			utils.WriteResponse(w, http.StatusCreated, account)	
		}
	}
}

func (h AccountHandler) MakeTransaction(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    accountId := vars["account_id"]
    customerId := vars["customer_id"]

    var request dto.TransactionRequest
    if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
        utils.WriteResponse(w, http.StatusBadRequest, err.Error())
    } else {
        request.AccountId = accountId
        request.CustomerId = customerId

        account, appError := h.service.MakeTransaction(request)

        if appError != nil {
            utils.WriteResponse(w, appError.Code, appError.AsMessage())
        } else {
            utils.WriteResponse(w, http.StatusOK, account)
        }
    }
}