package configs

import (
	"net/url"
	"slices"

	"example.com/tfgrid-kyc-service/internal/errors"
	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	MongoDB      MongoDB
	Server       Server
	Idenfy       Idenfy
	TFChain      TFChain
	Verification Verification
	IPLimiter    IPLimiter
	IDLimiter    IDLimiter
	Challenge    Challenge
	Log          Log
}

type MongoDB struct {
	URI          string `env:"MONGO_URI" env-default:"mongodb://localhost:27017"`
	DatabaseName string `env:"DATABASE_NAME" env-default:"tf-kyc-db"`
}
type Server struct {
	Port string `env:"PORT" env-default:"8080"`
}
type Idenfy struct {
	APIKey          string   `env:"IDENFY_API_KEY" env-required:"true"`
	APISecret       string   `env:"IDENFY_API_SECRET" env-required:"true"`
	BaseURL         string   `env:"IDENFY_BASE_URL" env-default:"https://ivs.idenfy.com"`
	CallbackSignKey string   `env:"IDENFY_CALLBACK_SIGN_KEY" env-required:"true"`
	WhitelistedIPs  []string `env:"IDENFY_WHITELISTED_IPS" env-separator:","`
	DevMode         bool     `env:"IDENFY_DEV_MODE" env-default:"false"`
	CallbackUrl     string   `env:"IDENFY_CALLBACK_URL" env-required:"false"`
	Namespace       string   `env:"IDENFY_NAMESPACE" env-default:""`
}
type TFChain struct {
	WsProviderURL string `env:"TFCHAIN_WS_PROVIDER_URL" env-default:"wss://tfchain.grid.tf"`
}
type Verification struct {
	SuspiciousVerificationOutcome string   `env:"VERIFICATION_SUSPICIOUS_VERIFICATION_OUTCOME" env-default:"APPROVED"`
	ExpiredDocumentOutcome        string   `env:"VERIFICATION_EXPIRED_DOCUMENT_OUTCOME" env-default:"REJECTED"`
	MinBalanceToVerifyAccount     uint64   `env:"VERIFICATION_MIN_BALANCE_TO_VERIFY_ACCOUNT" env-default:"10000000"`
	AlwaysVerifiedIDs             []string `env:"VERIFICATION_ALWAYS_VERIFIED_IDS" env-separator:","`
}
type IPLimiter struct {
	MaxTokenRequests int `env:"IP_LIMITER_MAX_TOKEN_REQUESTS" env-default:"4"`
	TokenExpiration  int `env:"IP_LIMITER_TOKEN_EXPIRATION" env-default:"1440"`
}
type IDLimiter struct {
	MaxTokenRequests int `env:"ID_LIMITER_MAX_TOKEN_REQUESTS" env-default:"4"`
	TokenExpiration  int `env:"ID_LIMITER_TOKEN_EXPIRATION" env-default:"1440"`
}
type Log struct {
	Debug bool `env:"DEBUG" env-default:"false"`
}
type Challenge struct {
	Window int64  `env:"CHALLENGE_WINDOW" env-default:"8"`
	Domain string `env:"CHALLENGE_DOMAIN" env-required:"true"`
}

func LoadConfig() (*Config, error) {
	cfg := &Config{}
	err := cleanenv.ReadEnv(cfg)
	if err != nil {
		return nil, errors.NewInternalError("error loading config", err)
	}
	cfg.Validate()
	return cfg, nil
}

// validate config
func (c *Config) Validate() error {
	// iDenfy base URL should be https://ivs.idenfy.com
	if c.Idenfy.BaseURL != "https://ivs.idenfy.com" {
		panic("invalid iDenfy base URL")
	}
	// CallbackUrl should be valid URL
	parsedCallbackUrl, err := url.ParseRequestURI(c.Idenfy.CallbackUrl)
	if err != nil {
		panic("invalid CallbackUrl")
	}
	// CallbackSignKey should not be empty
	if len(c.Idenfy.CallbackSignKey) < 16 {
		panic("CallbackSignKey should be at least 16 characters long")
	}
	// WsProviderURL should be valid URL and start with wss://
	if u, err := url.ParseRequestURI(c.TFChain.WsProviderURL); err != nil || u.Scheme != "wss" {
		panic("invalid WsProviderURL")
	}
	// domain should not be empty and same as domain in CallbackUrl
	if parsedCallbackUrl.Host != c.Challenge.Domain {
		panic("invalid Challenge Domain. It should be same as domain in CallbackUrl")
	}
	// Window should be greater than 2
	if c.Challenge.Window < 2 {
		panic("invalid Challenge Window. It should be greater than 2 otherwise it will be too short and verification can fail in slow networks")
	}
	// SuspiciousVerificationOutcome should be either APPROVED or REJECTED
	if !slices.Contains([]string{"APPROVED", "REJECTED"}, c.Verification.SuspiciousVerificationOutcome) {
		panic("invalid SuspiciousVerificationOutcome")
	}
	// ExpiredDocumentOutcome should be either APPROVED or REJECTED
	if !slices.Contains([]string{"APPROVED", "REJECTED"}, c.Verification.ExpiredDocumentOutcome) {
		panic("invalid ExpiredDocumentOutcome")
	}
	return nil
}
