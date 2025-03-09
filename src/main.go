package main

import (
	"github.com/titi0001/Microservices-API-in-Go/src/app"
	"github.com/titi0001/Microservices-API-in-Go/src/logger"
)

func main() {

	logger.Info("Starting the application...")
	app.Start()
}
