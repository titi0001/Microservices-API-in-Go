package app

import (
	"context"
	"net/http"
	"os"
	"strings"
	"sync"

	"github.com/gorilla/mux"
	"github.com/titi0001/Microservices-API-in-Go/src/domain"
	"github.com/titi0001/Microservices-API-in-Go/src/infrastructure/database"
	"github.com/titi0001/Microservices-API-in-Go/src/logger"
	"github.com/titi0001/Microservices-API-in-Go/src/service"
)

var (
	authServer       *http.Server
	authServerCancel context.CancelFunc
	authServerWg     sync.WaitGroup
)

func StartAuthServer() {
	authServerMutex.Lock()
	defer authServerMutex.Unlock()

	if authServer != nil {
		logger.Info("Auth server already running")
		return
	}

	authHost := os.Getenv("AUTH_LOCAL_HOST")
	if authHost == "" {
		logger.Fatal("AUTH_LOCAL_HOST environment variable not set")
	}

	router := mux.NewRouter()
	dbClient := database.GetClient()
	authRepositoryDb := domain.NewAuthRepositoryDb(dbClient)

	serviceURL := authHost
	if !strings.HasPrefix(serviceURL, "http://") {
		serviceURL = "http://" + serviceURL
	}

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

	authServer = &http.Server{
		Addr:    authHost,
		Handler: router,
	}

	var ctx context.Context
	ctx, authServerCancel = context.WithCancel(context.Background())

	authServerWg.Add(1)
	go func() {
		defer authServerWg.Done()

		logger.Info("Auth server starting on", logger.String("address", authHost))
		if err := authServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("Error starting auth server", logger.Any("error", err))
		}

		dbClient.Close()
		logger.Info("Auth server resources cleaned up")
	}()

	go func() {
		<-ctx.Done()

		shutdownCtx, cancel := context.WithTimeout(context.Background(), AuthServerShutdownTimeout)
		defer cancel()

		logger.Info("Shutting down auth server")
		if err := authServer.Shutdown(shutdownCtx); err != nil {
			logger.Error("Error shutting down auth server", logger.Any("error", err))
		}

		authServerMutex.Lock()
		authServer = nil
		authServerMutex.Unlock()

		logger.Info("Auth server shut down successfully")
	}()
}

func StopAuthServer() {
	authServerMutex.Lock()
	defer authServerMutex.Unlock()

	if authServer == nil {
		logger.Info("Auth server not running")
		return
	}
	if authServerCancel != nil {
		authServerCancel()
	}

	authServerWg.Wait()
}
