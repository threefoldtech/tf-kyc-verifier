// Use substarte client to get account free balance for development use
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
	free_balance, err := substrateClient.GetAccountBalance("5DFkH2fcqYecVHjfgAEfxgsJyoEg5Kd93JFihfpHDaNoWagJ")
	if err != nil {
		panic(err)
	}
	fmt.Println(free_balance)
}

type TFChainConfig struct {
	WsProviderURL string
}

// implement SubstrateConfig for config.TFChain
func (c *TFChainConfig) GetWsProviderURL() string {
	return c.WsProviderURL
}
