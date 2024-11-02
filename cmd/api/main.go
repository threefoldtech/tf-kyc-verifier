package main

import (
	"log"

	_ "example.com/tfgrid-kyc-service/api/docs"
	"example.com/tfgrid-kyc-service/internal/configs"
	"example.com/tfgrid-kyc-service/internal/logger"
	"example.com/tfgrid-kyc-service/internal/server"
)

//	@title			TFGrid KYC API
//	@version		0.1.0
//	@description	This is a KYC service for TFGrid.
//	@termsOfService	http://swagger.io/terms/

// @contact.name	Codescalers Egypt
// @contact.url	https://codescalers-egypt.com
// @contact.email	info@codescalers.com
// @BasePath		/
func main() {
	config, err := configs.LoadConfig()
	if err != nil {
		log.Fatal("Failed to load configuration:", err)
	}

	logger.Init(config.Log)
	log := logger.GetLogger()

	log.Debug("Configuration loaded successfully", logger.Fields{
		"config": config.GetPublicConfig(),
	})

	server, err := server.New(config, log)
	if err != nil {
		log.Fatal("Failed to create server:", logger.Fields{
			"error": err,
		})
	}

	log.Info("Starting server on port:", logger.Fields{
		"port": config.Server.Port,
	})
	server.Start()
}
