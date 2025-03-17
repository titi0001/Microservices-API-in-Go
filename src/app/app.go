package app

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/gorilla/mux"
	"github.com/jmoiron/sqlx"
	"github.com/joho/godotenv"
	"github.com/titi0001/Microservices-API-in-Go/src/domain"
	"github.com/titi0001/Microservices-API-in-Go/src/infrastructure/database"
	"github.com/titi0001/Microservices-API-in-Go/src/logger"
	"github.com/titi0001/Microservices-API-in-Go/src/service"
)

var (
	authServerMutex sync.Mutex
	authServiceURL  string
)

const (
	MainServerShutdownTimeout = 5 * time.Second
	AuthServerShutdownTimeout = 5 * time.Second
)

func Start() {
	err := godotenv.Load()
	if err != nil {
		logger.Fatal("Error loading .env file", logger.Any("error", err))
	}

	localHost := os.Getenv("LOCAL_HOST")
	authHost := os.Getenv("AUTH_LOCAL_HOST")

	if localHost == "" || authHost == "" {
		logger.Fatal("Required environment variables not set",
			logger.String("LOCAL_HOST", localHost),
			logger.String("AUTH_LOCAL_HOST", authHost))
	}

	authServiceURL = authHost
	if !strings.HasPrefix(authServiceURL, "http://") {
		authServiceURL = "http://" + authServiceURL
	}

	dbClient := database.GetClient()
	if dbClient == nil {
		logger.Fatal("Failed to connect to database")
	}

	var wg sync.WaitGroup
	wg.Add(2)

	authServer := setupAuthServer(authHost, authServiceURL, dbClient)
	go func() {
		defer wg.Done()
		logger.Info("Auth server starting on", logger.String("address", authHost))
		if err := authServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("Error starting auth server", logger.Any("error", err))
		}
	}()

	time.Sleep(200 * time.Millisecond)

	mainServer := setupMainServer(localHost, authServiceURL, dbClient)
	go func() {
		defer wg.Done()
		logger.Info("Main server starting on", logger.String("address", localHost))
		if err := mainServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("Error starting main server", logger.Any("error", err))
		}
	}()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	<-sigChan

	logger.Info("Received shutdown signal, stopping servers")

	shutdownServer(mainServer, "main server", MainServerShutdownTimeout)
	shutdownServer(authServer, "auth server", AuthServerShutdownTimeout)

	dbClient.Close()
	logger.Info("All servers shut down successfully")
}

func setupAuthServer(host string, serviceURL string, dbClient *sqlx.DB) *http.Server {
	router := mux.NewRouter()

	authRepositoryDb := domain.NewAuthRepositoryDb(dbClient)
	authService := service.NewAuthService(serviceURL, authRepositoryDb)
	authHandler := NewAuthHandler(authService)

	router.
		HandleFunc("/auth/login", authHandler.Login).
		Methods(http.MethodPost).
		Name("AuthLogin")

	router.
		HandleFunc("/auth/verify", authHandler.Verify).
		Methods(http.MethodGet).
		Name("VerifyToken")

	return &http.Server{
		Addr:         host,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}
}

func setupMainServer(host string, authServerURL string, dbClient *sqlx.DB) *http.Server {
	router := mux.NewRouter()

	customerRepositoryDb := domain.NewCustomerRepositoryDb(dbClient)
	accountRepositoryDB := domain.NewAccountRepositoryDb(dbClient)
	authRepositoryDb := domain.NewAuthRepositoryDb(dbClient)

	customerService := service.NewCustomerService(customerRepositoryDb)
	accountService := service.NewAccountService(accountRepositoryDB)
	authService := service.NewAuthService(authServerURL, authRepositoryDb)

	ch := CustomerHandler{service: customerService}
	ah := AccountHandler{service: accountService}
	auth := NewAuthHandler(authService)
	am := NewAuthMiddleware(authRepositoryDb)

	setupRoutes(router, ch, ah, auth, am)

	return &http.Server{
		Addr:         host,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}
}

func shutdownServer(server *http.Server, name string, timeout time.Duration) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		logger.Error("Error shutting down "+name, logger.Any("error", err))
	} else {
		logger.Info(name + " shut down successfully")
	}
}

func setupRoutes(router *mux.Router, ch CustomerHandler, ah AccountHandler, auth *AuthHandler, am AuthMiddleware) {
	publicRouter := router.PathPrefix("").Subrouter()

	publicRouter.
		HandleFunc("/auth/login", auth.Login).
		Methods(http.MethodPost).
		Name("AuthLogin")

	protectedRouter := router.PathPrefix("").Subrouter()
	protectedRouter.Use(am.authorizationHandler())

	protectedRouter.
		HandleFunc("/auth/verify", auth.Verify).
		Methods(http.MethodGet).
		Name("AuthVerify")

	protectedRouter.
		HandleFunc("/customers", ch.GetAllCustomers).
		Methods(http.MethodGet).
		Name("GetAllCustomers")

	protectedRouter.
		HandleFunc("/customers/{customer_id:[0-9]+}", ch.GetCustomer).
		Methods(http.MethodGet).
		Name("GetCustomer")

	protectedRouter.
		HandleFunc("/customers/{customer_id:[0-9]+}/account", ah.NewAccount).
		Methods(http.MethodPost).
		Name("NewAccount")

	protectedRouter.
		HandleFunc("/customers/{customer_id:[0-9]+}/account/{account_id:[0-9]+}", ah.MakeTransaction).
		Methods(http.MethodPost).
		Name("NewTransaction")
}

func GetAuthServiceURL() string {
	return authServiceURL
}
