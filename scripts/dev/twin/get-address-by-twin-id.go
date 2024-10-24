package main

import (
	"fmt"

	"example.com/tfgrid-kyc-service/internal/clients/substrate"
	"example.com/tfgrid-kyc-service/internal/configs"
)

func main() {
	config, err := configs.LoadConfig()
	if err != nil {
		panic(err)
	}
	substrateClient, err := substrate.New(config.TFChain)
	if err != nil {
		panic(err)
	}

	address, err := substrateClient.GetAddressByTwinID("41")
	if err != nil {
		panic(err)
	}
	fmt.Println(address)

}
