package service

import (
	"github.com/titi0001/Microservices-API-in-Go/api/dto"
	"github.com/titi0001/Microservices-API-in-Go/domain"
	"github.com/titi0001/Microservices-API-in-Go/domain/ports"
	"github.com/titi0001/Microservices-API-in-Go/errs"
	"github.com/titi0001/Microservices-API-in-Go/logger"
)

type DefaultAccountService struct {
	repo ports.AccountRepository
}

func NewAccountService(repo ports.AccountRepository) ports.AccountService {
	return &DefaultAccountService{repo: repo}
}


func (s *DefaultAccountService) NewAccount(req dto.NewAccountRequest) (*dto.NewAccountResponse, *errs.AppError) {
	account := domain.NewAccount(req.CustomerID, req.AccountType, req.Amount)
	savedAccount, err := s.repo.Save(account)
	if err != nil {
		logger.Error("Error saving new account", logger.Any("error", err))
		return nil, err
	}
	return savedAccount.ToNewAccountResponseDto(), nil
}


func (s *DefaultAccountService) MakeTransaction(req dto.TransactionRequest) (*dto.TransactionResponse, *errs.AppError) {

	account, err := s.repo.FindBy(req.AccountID)
	if err != nil {
		logger.Error("Error finding account", logger.String("account_id", req.AccountID), logger.Any("error", err))
		return nil, err
	}

	if req.IsTransactionTypeWithdrawal() && !account.CanWithdraw(req.Amount) {
		logger.Warn("Insufficient balance for withdrawal", logger.String("account_id", req.AccountID), logger.Float64("amount", req.Amount))
		return nil, errs.NewValidationError("Insufficient balance for withdrawal")
	}


	transaction := domain.Transaction{
		AccountID:       req.AccountID,
		Amount:          req.Amount,
		TransactionType: req.TransactionType,
		TransactionDate: req.TransactionDate,
	}

	savedTransaction, saveErr := s.repo.SaveTransaction(transaction)
	if saveErr != nil {
		logger.Error("Error saving transaction", logger.String("account_id", req.AccountID), logger.Any("error", saveErr))
		return nil, saveErr
	}

	response := savedTransaction.ToDto()
	return &response, nil
}