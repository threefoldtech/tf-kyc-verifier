package idenfy

type IdenfyConfig interface {
	GetBaseURL() string
	GetCallbackUrl() string
	GetNamespace() string
	GetDevMode() bool
	GetWhitelistedIPs() []string
	GetAPIKey() string
	GetAPISecret() string
	GetCallbackSignKey() string
}
