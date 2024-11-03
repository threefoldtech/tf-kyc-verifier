package main

import (
	"log"

	_ "github.com/threefoldtech/tf-kyc-verifier/api/docs"
	"github.com/threefoldtech/tf-kyc-verifier/internal/configs"
	"github.com/threefoldtech/tf-kyc-verifier/internal/logger"
	"github.com/threefoldtech/tf-kyc-verifier/internal/server"
)

//	@title			TFGrid KYC API
//	@version		0.3.0
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
	srvLogger := logger.GetLogger()

	srvLogger.Debug("Configuration loaded successfully", logger.Fields{
		"config": config.GetPublicConfig(),
	})

	server, err := server.New(config, srvLogger)
	if err != nil {
		srvLogger.Error("Failed to create server:", logger.Fields{
			"error": err,
		})
	}

	srvLogger.Info("Starting server on port:", logger.Fields{
		"port": config.Server.Port,
	})
	server.Start()
}
