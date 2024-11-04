package main

import (
	"fmt"
	"log"

	"github.com/threefoldtech/tf-kyc-verifier/internal/clients/substrate"
	"github.com/threefoldtech/tf-kyc-verifier/internal/logger"
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

	chainName, err := substrateClient.GetChainName()
	if err != nil {
		panic(err)
	}
	fmt.Println(chainName)

}

// implement logger.LoggerW for log.Logger
type LoggerW struct {
	*log.Logger
}

func (l *LoggerW) Debug(msg string, fields logger.Fields) {
	l.Println(msg)
}

func (l *LoggerW) Info(msg string, fields logger.Fields) {
	l.Println(msg)
}

func (l *LoggerW) Warn(msg string, fields logger.Fields) {
	l.Println(msg)
}

func (l *LoggerW) Error(msg string, fields logger.Fields) {
	l.Println(msg)
}

func (l *LoggerW) Fatal(msg string, fields logger.Fields) {
	l.Println(msg)
}

type TFChainConfig struct {
	WsProviderURL string
}

// implement SubstrateConfig for config.TFChain
func (c *TFChainConfig) GetWsProviderURL() string {
	return c.WsProviderURL
}
