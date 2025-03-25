package domain

import (
	"github.com/titi0001/Microservices-API-in-Go/api/dto"
)

const (
	Withdrawal = "withdrawal"
	Deposit    = "deposit"
)

type Transaction struct {
	TransactionID   string  `json:"transaction_id"`
	AccountID       string  `json:"account_id"`
	Amount          float64 `json:"amount"`
	TransactionType string  `json:"transaction_type"`
	TransactionDate string  `json:"transaction_date"`
}

func (t Transaction) IsWithdrawal() bool {
	return t.TransactionType == Withdrawal
}

func (t Transaction) ToDto() dto.TransactionResponse {
	return dto.TransactionResponse{
		TransactionID:   t.TransactionID,
		AccountID:       t.AccountID,
		NewBalance:      t.Amount, 
		TransactionType: t.TransactionType,
		TransactionDate: t.TransactionDate,
	}
}