package repository

import (
    "database/sql"
    "strconv"

    "github.com/jmoiron/sqlx"
    "github.com/titi0001/Microservices-API-in-Go/domain"
    "github.com/titi0001/Microservices-API-in-Go/domain/ports"
    "github.com/titi0001/Microservices-API-in-Go/errs"
    "github.com/titi0001/Microservices-API-in-Go/logger"
)

type AccountRepositoryDb struct {
    client *sqlx.DB
}

func NewAccountRepositoryDb(dbClient *sqlx.DB) AccountRepositoryDb {
    return AccountRepositoryDb{client: dbClient}
}

func (d AccountRepositoryDb) Save(a domain.Account) (*domain.Account, *errs.AppError) {
    sqlInsert := "INSERT INTO accounts (customer_id, opening_date, account_type, amount, status) VALUES (?, ?, ?, ?, ?)"
    result, err := d.client.Exec(sqlInsert, a.CustomerID, a.OpeningDate, a.AccountType, a.Amount, a.Status)
    if err != nil {
        logger.Error("Error creating new account", logger.Any("error", err))
        return nil, errs.NewUnexpectedError("Unexpected database error")
    }

    id, err := result.LastInsertId()
    if err != nil {
        logger.Error("Error getting last insert ID", logger.Any("error", err))
        return nil, errs.NewUnexpectedError("Unexpected database error")
    }

    a.AccountID = strconv.FormatInt(id, 10)
    return &a, nil
}

func (d AccountRepositoryDb) SaveTransaction(t domain.Transaction) (*domain.Transaction, *errs.AppError) {
    tx, err := d.client.Begin()
    if err != nil {
        logger.Error("Error starting transaction", logger.Any("error", err))
        return nil, errs.NewUnexpectedError("Unexpected database error")
    }

    result, err := tx.Exec(
        "INSERT INTO transactions (account_id, amount, transaction_type, transaction_date) VALUES (?, ?, ?, ?)",
        t.AccountID, t.Amount, t.TransactionType, t.TransactionDate,
    )
    if err != nil {
        if rollbackErr := tx.Rollback(); rollbackErr != nil {
            logger.Error("Error rolling back transaction", logger.Any("error", rollbackErr))
        }
        logger.Error("Error inserting transaction", logger.Any("error", err))
        return nil, errs.NewUnexpectedError("Unexpected database error")
    }

    updateQuery := "UPDATE accounts SET amount = amount + ? WHERE account_id = ?"
    if t.IsWithdrawal() {
        updateQuery = "UPDATE accounts SET amount = amount - ? WHERE account_id = ?"
    }
    _, err = tx.Exec(updateQuery, t.Amount, t.AccountID)
    if err != nil {
        if rollbackErr := tx.Rollback(); rollbackErr != nil {
            logger.Error("Error rolling back transaction", logger.Any("error", rollbackErr))
        }
        logger.Error("Error updating account balance", logger.Any("error", err))
        return nil, errs.NewUnexpectedError("Unexpected database error")
    }

    if err = tx.Commit(); err != nil {
        logger.Error("Error committing transaction", logger.Any("error", err))
        return nil, errs.NewUnexpectedError("Unexpected database error")
    }

    transactionID, err := result.LastInsertId()
    if err != nil {
        logger.Error("Error getting transaction ID", logger.Any("error", err))
        return nil, errs.NewUnexpectedError("Unexpected database error")
    }

    account, appErr := d.FindBy(t.AccountID)
    if appErr != nil {
        return nil, appErr
    }

    t.TransactionID = strconv.FormatInt(transactionID, 10)
    t.Amount = account.Amount
    return &t, nil
}

func (d AccountRepositoryDb) FindBy(accountID string) (*domain.Account, *errs.AppError) {
    sqlGetAccount := "SELECT account_id, customer_id, opening_date, account_type, amount, status FROM accounts WHERE account_id = ?"
    var account domain.Account
    err := d.client.Get(&account, sqlGetAccount, accountID)
    if err != nil {
        if err == sql.ErrNoRows {
            logger.Warn("Account not found", logger.String("account_id", accountID))
            return nil, errs.NewNotFoundError("Account not found")
        }
        logger.Error("Error fetching account", logger.Any("error", err))
        return nil, errs.NewUnexpectedError("Unexpected database error")
    }
    return &account, nil
}


var _ ports.AccountRepository = (*AccountRepositoryDb)(nil)