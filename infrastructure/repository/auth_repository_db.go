package repository

import (
	"database/sql"
	"strconv"

	"github.com/jmoiron/sqlx"
	"github.com/titi0001/Microservices-API-in-Go/domain"
	"github.com/titi0001/Microservices-API-in-Go/domain/ports"
	"github.com/titi0001/Microservices-API-in-Go/errs"
	"github.com/titi0001/Microservices-API-in-Go/infrastructure/utils"
	"github.com/titi0001/Microservices-API-in-Go/logger"
)

type AuthRepositoryDb struct {
	client *sqlx.DB
}

type RemoteAuthRepository struct {
	authService ports.AuthService
}

func NewAuthRepositoryDb(dbClient *sqlx.DB) AuthRepositoryDb {
	return AuthRepositoryDb{client: dbClient}
}

func NewRemoteAuthRepository(authService ports.AuthService) RemoteAuthRepository {
	return RemoteAuthRepository{authService: authService}
}

func (d AuthRepositoryDb) FindUser(username, password string) (*domain.User, *errs.AppError) {
	query := `SELECT username, role, customer_id, created_on 
              FROM users 
              WHERE username = ? AND password = ?`
	var user domain.User
	err := d.client.Get(&user, query, username, password)
	if err != nil {
		if err == sql.ErrNoRows {
			logger.Warn("Invalid credentials", logger.String("username", username))
			return nil, errs.NewAuthenticationError("Invalid credentials")
		}
		logger.Error("Error querying user", logger.Any("error", err))
		return nil, errs.NewUnexpectedError("Unexpected database error")
	}
	return &user, nil
}

func (d AuthRepositoryDb) SaveUser(user domain.User) (*domain.User, *errs.AppError) {
	query := `INSERT INTO users (username, password, role, customer_id, created_on) 
              VALUES (?, ?, ?, ?, ?)`
	result, err := d.client.Exec(query, user.Username, user.Password, user.Role, user.CustomerID, user.CreatedOn)
	if err != nil {
		logger.Error("Error saving new user", logger.Any("error", err))
		return nil, errs.NewUnexpectedError("Unexpected database error")
	}

	id, err := result.LastInsertId()
	if err != nil {
		logger.Error("Error getting last insert ID for new user", logger.Any("error", err))
		return nil, errs.NewUnexpectedError("Unexpected database error")
	}

	user.ID = strconv.FormatInt(id, 10)
	return &user, nil
}

func (d AuthRepositoryDb) VerifyPermission(role string, customerId string, routeName string, vars map[string]string) bool {
	if !d.verifyAdminRoute(role, routeName) {
		return false
	}
	if !d.verifyCustomerSpecificRoute(role, customerId, routeName, vars) {
		return false
	}
	return true
}

func (d AuthRepositoryDb) verifyAdminRoute(role string, routeName string) bool {
	adminOnlyRoutes := map[string]bool{
		"GetAllCustomers": true,
		"DeleteAccount":   true,
		"GetCustomers":    true,
		"CreateCustomer":  true,
		"DeleteCustomer":  true,
	}
	if adminOnlyRoutes[routeName] && role != "admin" {
		logger.Warn("Permission denied - admin route access attempt",
			logger.String("role", role),
			logger.String("routeName", routeName))
		return false
	}
	return true
}

func (d AuthRepositoryDb) verifyCustomerSpecificRoute(role string, customerId string, routeName string, vars map[string]string) bool {
	customerSpecificRoutes := map[string]bool{
		"GetCustomerById":         true,
		"UpdateCustomer":          true,
		"GetAccountsByCustomerId": true,
	}
	if !customerSpecificRoutes[routeName] {
		return true
	}
	if role == "admin" {
		return true
	}
	routeCustomerId, exists := vars["customer_id"]
	if !exists {
		return true
	}
	if customerId == "" {
		logger.Warn("Customer ID not found for user", logger.String("role", role))
		return false
	}
	if customerId != routeCustomerId {
		logger.Warn("Permission denied - customer data access attempt",
			logger.String("role", role),
			logger.String("requested_id", routeCustomerId),
			logger.String("user_id", customerId))
		return false
	}
	return true
}

func (r RemoteAuthRepository) FindUser(username, password string) (*domain.User, *errs.AppError) {
	return nil, errs.NewUnexpectedError("Method not implemented")
}

func (r RemoteAuthRepository) SaveUser(user domain.User) (*domain.User, *errs.AppError) {
	return nil, errs.NewUnexpectedError("Method not implemented")
}

func (r RemoteAuthRepository) VerifyPermission(role string, customerId string, routeName string, vars map[string]string) bool {
	url := utils.BuildVerifyURL("some-token", routeName, vars)
	logger.Info("Generated URL for remote auth", logger.String("url", url))
	return false
}