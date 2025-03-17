package app

import (
	"context"
	"errors"
	"net"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/mux"
	"github.com/titi0001/Microservices-API-in-Go/src/domain"
	"github.com/titi0001/Microservices-API-in-Go/src/infrastructure/database"
	"github.com/titi0001/Microservices-API-in-Go/src/logger"
	"github.com/titi0001/Microservices-API-in-Go/src/service"
)

const (
	AuthServerDialTimeout = 100 * time.Millisecond
)

var (
	authServer        *http.Server
	authServerCancel  context.CancelFunc
	authServerWg      sync.WaitGroup
	authServerStarted bool
)

func StartAuthServer() error {
	authServerMutex.Lock()
	defer authServerMutex.Unlock()

	if authServerStarted {
		logger.Info("Auth server already started or attempted to start")
		return nil
	}
	authServerStarted = true

	if authServer != nil {
		logger.Info("Auth server instance already exists")
		return nil
	}

	authHost := os.Getenv("AUTH_LOCAL_HOST")
	if authHost == "" {
		logger.Error("AUTH_LOCAL_HOST environment variable not set")
		return errors.New("AUTH_LOCAL_HOST environment variable not set")
	}

	conn, err := net.DialTimeout("tcp", authHost, AuthServerDialTimeout)
	if err == nil {
		conn.Close()
		logger.Info("Auth server port already in use, assuming server is running externally")
		return nil
	}

	router := mux.NewRouter()
	dbClient := database.GetClient()
	if dbClient == nil {
		logger.Error("Failed to connect to database")
		return errors.New("failed to connect to database")
	}
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
		Addr:         authHost,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
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
		if dbClient != nil {
			dbClient.Close()
		}
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
	return nil
}

func StopAuthServer() {
	authServerMutex.Lock()
	if authServer == nil {
		authServerMutex.Unlock()
		logger.Info("Auth server not running")
		return
	}
	cancel := authServerCancel
	authServerMutex.Unlock()

	if cancel != nil {
		cancel()
	}
	authServerWg.Wait()
	logger.Info("Auth server stopped completely")
}

func IsAuthServerRunning() bool {
	authHost := os.Getenv("AUTH_LOCAL_HOST")
	if authHost == "" {
		return false
	}

	conn, err := net.DialTimeout("tcp", authHost, AuthServerDialTimeout)
	if err != nil {
		return false
	}
	conn.Close()
	return true
}
