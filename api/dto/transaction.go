package dto

import (
	"github.com/titi0001/Microservices-API-in-Go/errs"
)


const (
	Withdrawal = "withdrawal"
	Deposit    = "deposit"
)

type TransactionRequest struct {
	AccountID       string  `json:"account_id"`
	Amount          float64 `json:"amount"`
	TransactionType string  `json:"transaction_type"`
	TransactionDate string  `json:"transaction_date"`
	CustomerID      string  `json:"-"` 
}

func (r TransactionRequest) IsTransactionTypeWithdrawal() bool {
	return r.TransactionType == Withdrawal
}

func (r TransactionRequest) IsTransactionTypeDeposit() bool {
	return r.TransactionType == Deposit
}

func (r TransactionRequest) Validate() *errs.AppError {
	if r.AccountID == "" {
		return errs.NewValidationError("Account ID is required")
	}
	if r.CustomerID == "" {
		return errs.NewValidationError("Customer ID is required")
	}
	if r.Amount <= 0 {
		return errs.NewValidationError("Amount must be greater than zero")
	}
	if !r.IsTransactionTypeWithdrawal() && !r.IsTransactionTypeDeposit() {
		return errs.NewValidationError("Transaction type must be 'deposit' or 'withdrawal'")
	}
	return nil
}


type TransactionResponse struct {
	TransactionID   string  `json:"transaction_id"`
	AccountID       string  `json:"account_id"`
	NewBalance      float64 `json:"new_balance"` 
	TransactionType string  `json:"transaction_type"`
	TransactionDate string  `json:"transaction_date"`
}