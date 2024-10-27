// Use substarte client to get account free balance for development use
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
	free_balance, err := substrateClient.GetAccountBalance("5DFkH2fcqYecVHjfgAEfxgsJyoEg5Kd93JFihfpHDaNoWagJ")
	if err != nil {
		panic(err)
	}
	fmt.Println(free_balance)
}
