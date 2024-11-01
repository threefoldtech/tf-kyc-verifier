package main

import (
	"fmt"
	"log"

	"example.com/tfgrid-kyc-service/internal/clients/substrate"
)

func main() {
	config := &TFChainConfig{
		WsProviderURL: "wss://tfchain.dev.grid.tf",
	}

	logger := &LoggerW{log.Default()}
	substrateClient, err := substrate.New(config, logger)
	if err != nil {
		panic(err)
	}

	address, err := substrateClient.GetAddressByTwinID("41")
	if err != nil {
		panic(err)
	}
	fmt.Println(address)

}

// implement logger.LoggerW for log.Logger
type LoggerW struct {
	*log.Logger
}

func (l *LoggerW) Debug(msg string, fields map[string]interface{}) {
	l.Println(msg)
}

func (l *LoggerW) Info(msg string, fields map[string]interface{}) {
	l.Println(msg)
}

func (l *LoggerW) Warn(msg string, fields map[string]interface{}) {
	l.Println(msg)
}

func (l *LoggerW) Error(msg string, fields map[string]interface{}) {
	l.Println(msg)
}

func (l *LoggerW) Fatal(msg string, fields map[string]interface{}) {
	l.Println(msg)
}

type TFChainConfig struct {
	WsProviderURL string
}

// implement SubstrateConfig for configs.TFChain
func (c *TFChainConfig) GetWsProviderURL() string {
	return c.WsProviderURL
}
