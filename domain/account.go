package domain

import (
	"time"

	"github.com/titi0001/Microservices-API-in-Go/api/dto"
)

type Account struct {
	AccountID   string  `json:"account_id"`
	CustomerID  string  `json:"customer_id"`
	OpeningDate string  `json:"opening_date"`
	AccountType string  `json:"account_type"`
	Amount      float64 `json:"amount"`
	Status      string  `json:"status"`
}

func NewAccount(customerID, accountType string, amount float64) Account {
	return Account{
		AccountID:   "",
		CustomerID:  customerID,
		OpeningDate: time.Now().Format("2006-01-02 15:04:05"),
		AccountType: accountType,
		Amount:      amount,
		Status:      "1",
	}
}

func (a Account) ToNewAccountResponseDto() *dto.NewAccountResponse {
	return &dto.NewAccountResponse{AccountID: a.AccountID}
}

func (a Account) CanWithdraw(amount float64) bool {
	return a.Amount >= amount
}
