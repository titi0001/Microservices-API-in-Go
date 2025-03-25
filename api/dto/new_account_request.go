package dto

import (
	"strings"

	"github.com/titi0001/Microservices-API-in-Go/errs"
)

type NewAccountRequest struct {
	CustomerID  string  `json:"customer_id"`
	AccountType string  `json:"account_type"`
	Amount      float64 `json:"amount"`
}

func (r NewAccountRequest) Validate() *errs.AppError {
	if r.CustomerID == "" {
		return errs.NewValidationError("Customer ID is required")
	}
	if r.Amount < 5000 {
		return errs.NewValidationError("Initial deposit must be at least 5000.00")
	}
	lowerAccountType := strings.ToLower(r.AccountType)
	if lowerAccountType != "saving" && lowerAccountType != "checking" {
		return errs.NewValidationError("Account type must be 'saving' or 'checking'")
	}
	return nil
}