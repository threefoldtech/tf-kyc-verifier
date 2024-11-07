package main

import (
	"fmt"
	"log/slog"

	"github.com/threefoldtech/tf-kyc-verifier/internal/clients/substrate"
)

func main() {
	config := &TFChainConfig{
		WsProviderURL: "wss://tfchain.dev.grid.tf",
	}

	logger := slog.Default()
	substrateClient, err := substrate.New(config, logger)
	if err != nil {
		panic(err)
	}

	address, err := substrateClient.GetAddressByTwinID(41)
	if err != nil {
		panic(err)
	}
	fmt.Println(address)

}

type TFChainConfig struct {
	WsProviderURL string
}

// implement SubstrateConfig for config.TFChain
func (c *TFChainConfig) GetWsProviderURL() string {
	return c.WsProviderURL
}
