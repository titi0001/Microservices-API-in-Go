package api

import (
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/jmoiron/sqlx"
	"github.com/titi0001/Microservices-API-in-Go/domain/ports"
	"github.com/titi0001/Microservices-API-in-Go/domain/service"
	"github.com/titi0001/Microservices-API-in-Go/infrastructure/repository"
)

func SetupAuthServer(host, serviceURL string, dbClient *sqlx.DB) *http.Server {
	router := mux.NewRouter()

	authRepo := repository.NewAuthRepositoryDb(dbClient)
	authService := service.NewAuthService(serviceURL, authRepo)
	authHandler := NewAuthHandler(authService)

	router.HandleFunc("/auth/login", authHandler.Login).Methods(http.MethodPost).Name("AuthLogin")
	router.HandleFunc("/auth/verify", authHandler.Verify).Methods(http.MethodGet).Name("VerifyToken")

	return &http.Server{
		Addr:         host,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}
}

func SetupMainServer(host, authServerURL string, dbClient *sqlx.DB) *http.Server {
	router := mux.NewRouter()


	customerRepo := repository.NewCustomerRepositoryDb(dbClient)
	accountRepo := repository.NewAccountRepositoryDb(dbClient)
	authRepo := repository.NewAuthRepositoryDb(dbClient)


	customerService := service.NewCustomerService(customerRepo)
	accountService := service.NewAccountService(accountRepo)
	authService := service.NewAuthService(authServerURL, authRepo)


	authMiddleware := NewAuthMiddleware(authRepo)

	setupRoutes(router, customerService, accountService, authService, authMiddleware)

	return &http.Server{
		Addr:         host,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}
}


func setupRoutes(
	router *mux.Router,
	customerService ports.CustomerService,
	accountService ports.AccountService,
	authService ports.AuthService,
	authMiddleware *AuthMiddleware,
) {

	publicRouter := router.PathPrefix("").Subrouter()
	publicRouter.HandleFunc("/auth/login", NewAuthHandler(authService).Login).Methods(http.MethodPost).Name("AuthLogin")

	protectedRouter := router.PathPrefix("").Subrouter()
	protectedRouter.Use(authMiddleware.AuthorizationHandler())

	protectedRouter.HandleFunc("/auth/verify", NewAuthHandler(authService).Verify).Methods(http.MethodGet).Name("AuthVerify")
	protectedRouter.HandleFunc("/customers", NewCustomerHandler(customerService).GetAllCustomers).Methods(http.MethodGet).Name("GetAllCustomers")
	protectedRouter.HandleFunc("/customers/{customer_id:[0-9]+}", NewCustomerHandler(customerService).GetCustomer).Methods(http.MethodGet).Name("GetCustomer")
	protectedRouter.HandleFunc("/customers/{customer_id:[0-9]+}/account", NewAccountHandler(accountService).NewAccount).Methods(http.MethodPost).Name("NewAccount")
	protectedRouter.HandleFunc("/customers/{customer_id:[0-9]+}/account/{account_id:[0-9]+}", NewAccountHandler(accountService).MakeTransaction).Methods(http.MethodPost).Name("NewTransaction")
	protectedRouter.HandleFunc("/permissions", NewPermissionsHandler(authService).GetRolePermissions).Methods(http.MethodGet).Name("GetRolePermissions")
}