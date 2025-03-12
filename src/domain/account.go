package domain

import (
	"github.com/titi0001/Microservices-API-in-Go/src/dto"
	"github.com/titi0001/Microservices-API-in-Go/src/errs"
)

const dbTSLayout = "2006-01-02 15:04:05"

type Account struct {
	AccountId   string
	CustomerId  string
	OpeningDate string
	AccountType string
	Amount      float64
	Status      string
}

func (a Account) ToNewAccountResponseDto() *dto.NewAccountResponse {
	return &dto.NewAccountResponse{AccountId: a.AccountId}
}

type AccountRepository interface {
	Save(Account) (*Account, *errs.AppError)
	SaveTransaction(transaction Transaction) (*Transaction, *errs.AppError)
	FindBy(accountId string) (*Account, *errs.AppError)
}

func (a Account) CanWithdraw(amount float64) bool {
	return a.Amount < amount

}

func NewAccount(customerId, accountType string, amount float64) Account {
	return Account{
		CustomerId:  customerId,
		OpeningDate: dbTSLayout,
		AccountType: accountType,
		Amount:      amount,
		Status:      "1",
	}
}
