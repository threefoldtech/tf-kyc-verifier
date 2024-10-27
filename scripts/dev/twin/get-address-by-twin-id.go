package main

import (
	"fmt"

	"example.com/tfgrid-kyc-service/internal/clients/substrate"
	"example.com/tfgrid-kyc-service/internal/configs"
	"example.com/tfgrid-kyc-service/internal/logger"
)

func main() {
	config, err := configs.LoadConfig()
	if err != nil {
		panic(err)
	}
	logger.Init(config.Log)
	logger := logger.GetLogger()
	substrateClient, err := substrate.New(config.TFChain, logger)
	if err != nil {
		panic(err)
	}

	address, err := substrateClient.GetAddressByTwinID("41")
	if err != nil {
		panic(err)
	}
	fmt.Println(address)

}
