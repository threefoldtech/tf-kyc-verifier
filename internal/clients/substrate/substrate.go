package substrate

import (
	"fmt"
	"math/big"
	"strconv"
	"sync"

	"example.com/tfgrid-kyc-service/internal/configs"

	// use tfchain go client

	tfchain "github.com/threefoldtech/tfchain/clients/tfchain-client-go"
)

type Substrate struct {
	api *tfchain.Substrate
	mu  sync.Mutex // TODO: Check if SubstrateAPI is thread safe
}

func New(config configs.TFChain) (*Substrate, error) {
	mgr := tfchain.NewManager(config.WsProviderURL)
	api, err := mgr.Substrate()
	if err != nil {
		return nil, fmt.Errorf("substrate connection error: failed to initialize Substrate client: %w", err)
	}

	c := &Substrate{
		api: api,
		mu:  sync.Mutex{},
	}
	return c, nil
}

func (c *Substrate) GetAccountBalance(address string) (*big.Int, error) {
	pubkeyBytes, err := tfchain.FromAddress(address)
	if err != nil {
		return nil, fmt.Errorf("failed to decode ss58 address: %w", err)
	}
	accountID := tfchain.AccountID(pubkeyBytes)
	balance, err := c.api.GetBalance(accountID)
	if err != nil {
		return nil, fmt.Errorf("failed to get balance: %w", err)
	}

	return balance.Free.Int, nil
}

func (c *Substrate) GetAddressByTwinID(twinID string) (string, error) {
	twinIDUint32, err := strconv.ParseUint(twinID, 10, 32)
	if err != nil {
		return "", fmt.Errorf("failed to parse twin ID: %w", err)
	}
	twin, err := c.api.GetTwin(uint32(twinIDUint32))
	if err != nil {
		return "", fmt.Errorf("failed to get twin: %w", err)
	}
	return twin.Account.String(), nil
}
