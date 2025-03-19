package service

import (
	"testing"

	"github.com/golang/mock/gomock"
	realDomain "github.com/titi0001/Microservices-API-in-Go/src/domain"
	"github.com/titi0001/Microservices-API-in-Go/src/dto"
	"github.com/titi0001/Microservices-API-in-Go/src/errs"
	"github.com/titi0001/Microservices-API-in-Go/src/mocks/domain"
)

func Test_should_return_a_validation_error_response_when_the_request_is_not_validated(t *testing.T) {
	request := dto.NewAccountRequest{
		CustomerId:  "100",
		AccountType: "saving",
		Amount:      0,
	}

	service := NewAccountService(nil)

	_, appError := service.NewAccount(request)
	if appError == nil {
		t.Error("failed  while testing the new account validation")
	}
}

var (
	mockRepo *domain.MockAccountRepository
	service  AccountService
)

func setup(t *testing.T) func() {
	ctrl := gomock.NewController(t)
	mockRepo = domain.NewMockAccountRepository(ctrl)
	service = NewAccountService(mockRepo)

	return func() {
		service = nil
		defer ctrl.Finish()
	}

}

func Test_should_return_an_error_from_the_server_side_if_the_new_account_cannot_be_created(t *testing.T) {

	teardown := setup(t)
	defer teardown()

	req := dto.NewAccountRequest{
		CustomerId:  "100",
		AccountType: "saving",
		Amount:      6000,
	}

	account := realDomain.Account{
		CustomerId:  req.CustomerId,
		OpeningDate: dbTSLayout,
		AccountType: req.AccountType,
		Amount:      req.Amount,
		Status:      "1",
	}

	mockRepo.EXPECT().Save(account).Return(nil, errs.NewUnexpectedError("Unexpected database error"))

	_, appError := service.NewAccount(req)
	if appError == nil {
		t.Error("Test failed while Validating error for new account")
	}
}

func Test_should_return_new_account_response_when_a_new_account_is_saved_successfully(t *testing.T) {

	teardown := setup(t)
	defer teardown()

	req := dto.NewAccountRequest{
		CustomerId:  "100",
		AccountType: "saving",
		Amount:      6000,
	}

	account := realDomain.Account{
		CustomerId:  req.CustomerId,
		OpeningDate: dbTSLayout,
		AccountType: req.AccountType,
		Amount:      req.Amount,
		Status:      "1",
	}

	accountWithId := account
	accountWithId.AccountId = "201"
	mockRepo.EXPECT().Save(account).Return(&accountWithId, nil)

	newAccount, appError := service.NewAccount(req)
	if appError != nil {
		t.Error("Test failed while creating new account")
	}
	if newAccount.AccountId != accountWithId.AccountId {
		t.Error("Failed  while mathching new account id")
	}

}
