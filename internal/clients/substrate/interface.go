package substrate

type SubstrateConfig interface {
	GetWsProviderURL() string
}
