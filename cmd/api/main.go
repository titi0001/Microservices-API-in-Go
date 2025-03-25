package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/joho/godotenv"
	"github.com/titi0001/Microservices-API-in-Go/api"
	"github.com/titi0001/Microservices-API-in-Go/infrastructure/database"
	"github.com/titi0001/Microservices-API-in-Go/logger"
)

const (
	mainServerShutdownTimeout = 5 * time.Second
	authServerShutdownTimeout = 5 * time.Second
)

func main() {
	if err := godotenv.Load(); err != nil {
		logger.Fatal("Error loading .env file", logger.Any("error", err))
	}

	localHost := os.Getenv("LOCAL_HOST")
	authHost := os.Getenv("AUTH_LOCAL_HOST")
	if localHost == "" || authHost == "" {
		logger.Fatal("Required environment variables not set",
			logger.String("local_host", localHost),
			logger.String("auth_host", authHost))
	}

	authServiceURL := authHost
	if !strings.HasPrefix(authServiceURL, "http://") {
		authServiceURL = "http://" + authServiceURL
	}

	dbClient := database.GetClient()
	if dbClient == nil {
		logger.Fatal("Failed to initialize database client")
	}
	defer dbClient.Close()

	var wg sync.WaitGroup
	wg.Add(2)

	authServer := api.SetupAuthServer(authHost, authServiceURL, dbClient)
	go startServer(authServer, authHost, "auth server", &wg)

	time.Sleep(200 * time.Millisecond)

	mainServer := api.SetupMainServer(localHost, authServiceURL, dbClient)
	go startServer(mainServer, localHost, "main server", &wg)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	<-sigChan

	logger.Info("Received shutdown signal, stopping servers")
	shutdownServer(mainServer, "main server", mainServerShutdownTimeout)
	shutdownServer(authServer, "auth server", authServerShutdownTimeout)

	wg.Wait()
	logger.Info("All servers shut down successfully")
}

func startServer(server *http.Server, address, name string, wg *sync.WaitGroup) {
	defer wg.Done()
	logger.Info("Starting server", logger.String("name", name), logger.String("address", address))
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		logger.Error("Error starting server", logger.String("name", name), logger.Any("error", err))
	}
}

func shutdownServer(server *http.Server, name string, timeout time.Duration) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		logger.Error("Error shutting down server", logger.String("name", name), logger.Any("error", err))
		return
	}
	logger.Info("Server shut down successfully", logger.String("name", name))
}