package domain

import (
	"database/sql"
	"net/url"
	"reflect"
	"strconv"

	"github.com/jmoiron/sqlx"
	"github.com/titi0001/Microservices-API-in-Go/src/errs"
	"github.com/titi0001/Microservices-API-in-Go/src/logger"
)

type RemoteAuthRepository struct {
	authService AuthService
}

type AuthRepositoryDb struct {
	client *sqlx.DB
}

func (d AuthRepositoryDb) FindUser(username, password string) (*User, *errs.AppError) {
    query := `SELECT username, role, customer_id, created_on 
              FROM users 
              WHERE username = ? AND password = ?`

    var user User
    err := d.client.Get(&user, query, username, password)

    if err != nil {
        if err == sql.ErrNoRows {
            logger.Warn("Invalid credentials", logger.String("username", username))
            return nil, errs.NewAuthenticationError("Invalid credentials")
        }
        logger.Error("Error querying user: " + err.Error())
        return nil, errs.NewUnexpectedError("Database error")
    }

    return &user, nil
}

func (d AuthRepositoryDb) VerifyPermission(role string, customerId interface{}, routeName string, vars map[string]string) bool {
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

func (d AuthRepositoryDb) verifyCustomerSpecificRoute(role string, customerId interface{}, routeName string, vars map[string]string) bool {
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
	if customerId == nil {
		logger.Warn("Customer ID not found for user", logger.String("role", role))
		return false
	}

	tokenCustomerIdStr, err := d.convertToString(customerId)
	if err != nil {
		logger.Error("Failed to convert customer ID", logger.String("error", err.Message))
		return false
	}
	if tokenCustomerIdStr != routeCustomerId {
		logger.Warn("Permission denied - customer data access attempt",
			logger.String("role", role),
			logger.String("requested_id", routeCustomerId),
			logger.String("user_id", tokenCustomerIdStr))
		return false
	}
	return true
}

func (d AuthRepositoryDb) convertToString(value interface{}) (string, *errs.AppError) {
	switch v := value.(type) {
	case float64:
		return strconv.FormatFloat(v, 'f', 0, 64), nil
	case int64:
		return strconv.FormatInt(v, 10), nil
	case int:
		return strconv.Itoa(v), nil
	case string:
		return v, nil
	default:
		typeInfo := reflect.TypeOf(value).String()
		logger.Error("Unexpected customer ID type: " + typeInfo)
		return "", errs.NewUnexpectedError("Unexpected customer ID type")
	}
}

func (r RemoteAuthRepository) FindUser(username, password string) (*User, *errs.AppError) {
	return nil, errs.NewUnexpectedError("Method not implemented")
}

func (r RemoteAuthRepository) VerifyPermission(role string, customerId interface{}, routeName string, vars map[string]string) bool {
	return false
}

func BuildVerifyUrl(token string, routeName string, vars map[string]string) string {
	u := url.URL{Host: "localhost:8181", Path: "/auth/verify", Scheme: "http"}
	q := u.Query()
	q.Add("token", token)
	q.Add("routeName", routeName)
	for k, v := range vars {
		q.Add(k, v)
	}
	u.RawQuery = q.Encode()
	return u.String()
}

func NewAuthRepository(authService AuthService) RemoteAuthRepository {
	return RemoteAuthRepository{
		authService: authService,
	}
}

func NewAuthRepositoryDb(dbClient *sqlx.DB) AuthRepositoryDb {
	return AuthRepositoryDb{
		client: dbClient,
	}
}
