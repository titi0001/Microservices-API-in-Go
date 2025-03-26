package repository

import (
	"database/sql"
	"encoding/json"
	"strconv"
	"strings"

	"github.com/jmoiron/sqlx"
	"github.com/titi0001/Microservices-API-in-Go/domain"
	"github.com/titi0001/Microservices-API-in-Go/errs"
	"github.com/titi0001/Microservices-API-in-Go/infrastructure/utils"
	"github.com/titi0001/Microservices-API-in-Go/logger"
)

type AuthRepositoryDb struct {
	client *sqlx.DB
}

type RemoteAuthRepository struct {
	authService domain.AuthService
}

func NewAuthRepositoryDb(dbClient *sqlx.DB) AuthRepositoryDb {
	return AuthRepositoryDb{client: dbClient}
}

func NewRemoteAuthRepository(authService domain.AuthService) RemoteAuthRepository {
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
		if strings.Contains(err.Error(), "Duplicate entry") && strings.Contains(err.Error(), "users.PRIMARY") {
			logger.Error("User already exists", logger.String("username", user.Username))
			return nil, errs.NewValidationError("User with username " + user.Username + " already exists")
		}
		logger.Error("Error saving new user", logger.Any("error", err))
		return nil, errs.NewUnexpectedError("Unexpected database error")
	}

	id, err := result.LastInsertId()
	if err != nil {
		logger.Error("Error getting last insert ID", logger.Any("error", err))
		return nil, errs.NewUnexpectedError("Unexpected database error")
	}

	user.ID = strconv.FormatInt(id, 10)
	return &user, nil
}

func (d AuthRepositoryDb) VerifyPermission(role, customerID, routeName string, vars map[string]string) bool {
	if !d.verifyAdminRoute(role, routeName) {
		return false
	}
	if !d.verifyCustomerSpecificRoute(role, customerID, routeName, vars) {
		return false
	}
	return true
}

func (d AuthRepositoryDb) verifyAdminRoute(role, routeName string) bool {
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

func (d AuthRepositoryDb) verifyCustomerSpecificRoute(role, customerID, routeName string, vars map[string]string) bool {
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
	routeCustomerID, exists := vars["customer_id"]
	if !exists || customerID == "" {
		logger.Warn("Customer ID not found or invalid", logger.String("role", role))
		return false
	}
	if customerID != routeCustomerID {
		logger.Warn("Permission denied - customer data access attempt",
			logger.String("role", role),
			logger.String("requested_id", routeCustomerID),
			logger.String("user_id", customerID))
		return false
	}
	return true
}

func (d AuthRepositoryDb) SaveRefreshToken(refreshToken string) *errs.AppError {
	query := "INSERT INTO refresh_token_store (refresh_token) VALUES (?)"
	_, err := d.client.Exec(query, refreshToken)
	if err != nil {
		logger.Error("Error saving refresh token", logger.Any("error", err))
		return errs.NewUnexpectedError("Unexpected database error")
	}
	return nil
}

func (d AuthRepositoryDb) VerifyRefreshToken(refreshToken string) (bool, *errs.AppError) {
	var exists int
	query := "SELECT COUNT(*) FROM refresh_token_store WHERE refresh_token = ?"
	err := d.client.Get(&exists, query, refreshToken)
	if err != nil {
		logger.Error("Error verifying refresh token", logger.Any("error", err))
		return false, errs.NewUnexpectedError("Unexpected database error")
	}
	return exists > 0, nil
}

func (d AuthRepositoryDb) DeleteRefreshToken(refreshToken string) (int64, *errs.AppError) {
	result, err := d.client.Exec("DELETE FROM refresh_token_store WHERE refresh_token = ?", refreshToken)
	if err != nil {
		logger.Error("Error deleting refresh token", logger.Any("error", err))
		return 0, errs.NewUnexpectedError("Unexpected database error")
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		logger.Error("Error getting rows affected", logger.Any("error", err))
		return 0, errs.NewUnexpectedError("Unexpected database error")
	}
	return rowsAffected, nil
}

func (r RemoteAuthRepository) FindUser(username, password string) (*domain.User, *errs.AppError) {
	loginData := map[string]string{
		"username": username,
		"password": password,
	}
	body, err := json.Marshal(loginData)
	if err != nil {
		logger.Error("Error marshaling login data", logger.Any("error", err))
		return nil, errs.NewUnexpectedError("Unexpected error preparing login request")
	}

	response, appErr := r.authService.RemoteLogin(strings.NewReader(string(body)))
	if appErr != nil {
		logger.Error("Error during remote login", logger.Any("error", appErr))
		return nil, appErr
	}

	var user domain.User
	if err := json.Unmarshal(response, &user); err != nil {
		logger.Error("Error unmarshaling remote login response", logger.Any("error", err))
		return nil, errs.NewUnexpectedError("Unexpected error parsing remote response")
	}

	return &user, nil
}

func (r RemoteAuthRepository) SaveUser(user domain.User) (*domain.User, *errs.AppError) {
	return nil, errs.NewUnexpectedError("User creation not supported via remote auth service")
}

func (r RemoteAuthRepository) VerifyPermission(role, customerID, routeName string, vars map[string]string) bool {
	url := utils.BuildVerifyURL("some-token", routeName, vars)
	logger.Info("Generated URL for remote auth", logger.String("url", url))
	isAuthorized, appErr := r.authService.RemoteIsAuthorized("some-token", routeName, vars)
	if appErr != nil {
		logger.Error("Error verifying permission remotely", logger.Any("error", appErr))
		return false
	}
	return isAuthorized
}

func (r RemoteAuthRepository) DeleteRefreshToken(refreshToken string) (int64, *errs.AppError) {
	return 0, errs.NewUnexpectedError("Delete refresh token not supported via remote auth service")
}
