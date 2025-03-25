package ports

import (
    "github.com/titi0001/Microservices-API-in-Go/domain"
    "github.com/titi0001/Microservices-API-in-Go/errs"
)

type AccountRepository interface {
    Save(account domain.Account) (*domain.Account, *errs.AppError)
    SaveTransaction(transaction domain.Transaction) (*domain.Transaction, *errs.AppError)
    FindBy(accountID string) (*domain.Account, *errs.AppError)
}