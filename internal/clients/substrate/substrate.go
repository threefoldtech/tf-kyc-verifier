/*
Package substrate contains the Substrate client for the application.
This layer is responsible for interacting with the Substrate API. It wraps the tfchain go client and provide basic operations.
*/
package substrate

import (
	"fmt"
	"strconv"

	"github.com/threefoldtech/tf-kyc-verifier/internal/logger"
	tfchain "github.com/threefoldtech/tfchain/clients/tfchain-client-go"
)

type WsProviderURLGetter interface {
	GetWsProviderURL() string
}

type SubstrateClient interface {
	GetChainName() (string, error)
	GetAddressByTwinID(twinID string) (string, error)
	GetAccountBalance(address string) (uint64, error)
}

type Substrate struct {
	api    *tfchain.Substrate
	logger logger.Logger
}

func New(config WsProviderURLGetter, logger logger.Logger) (*Substrate, error) {
	mgr := tfchain.NewManager(config.GetWsProviderURL())
	api, err := mgr.Substrate()
	if err != nil {
		return nil, fmt.Errorf("initializing Substrate client: %w", err)
	}

	c := &Substrate{
		api:    api,
		logger: logger,
	}
	return c, nil
}

func (c *Substrate) GetAccountBalance(address string) (uint64, error) {
	pubkeyBytes, err := tfchain.FromAddress(address)
	if err != nil {
		return 0, fmt.Errorf("decoding ss58 address: %w", err)
	}
	accountID := tfchain.AccountID(pubkeyBytes)
	balance, err := c.api.GetBalance(accountID)
	if err != nil {
		if err.Error() == "account not found" {
			return 0, nil
		}
		return 0, fmt.Errorf("getting account balance: %w", err)
	}

	return balance.Free.Uint64(), nil
}

func (c *Substrate) GetAddressByTwinID(twinID string) (string, error) {
	twinIDUint32, err := strconv.ParseUint(twinID, 10, 32)
	if err != nil {
		return "", fmt.Errorf("parsing twin ID: %w", err)
	}
	twin, err := c.api.GetTwin(uint32(twinIDUint32))
	if err != nil {
		return "", fmt.Errorf("getting twin from tfchain: %w", err)
	}
	return twin.Account.String(), nil
}

// get chain name from ws provider url
func (c *Substrate) GetChainName() (string, error) {
	api, _, err := c.api.GetClient()
	if err != nil {
		return "", fmt.Errorf("getting substrate inner client: %w", err)
	}
	chain, err := api.RPC.System.Chain()
	if err != nil {
		return "", fmt.Errorf("getting chain name: %w", err)
	}
	return string(chain), nil
}
