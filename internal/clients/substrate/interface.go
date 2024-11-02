package substrate

import "math/big"

type SubstrateConfig interface {
	GetWsProviderURL() string
}

type SubstrateClient interface {
	GetChainName() (string, error)
	GetAddressByTwinID(twinID string) (string, error)
	GetAccountBalance(address string) (*big.Int, error)
}
