package configs

import (
	"example.com/tfgrid-kyc-service/internal/errors"
	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	MongoDB         MongoDB
	Server          Server
	Idenfy          Idenfy
	TFChain         TFChain
	Verification    Verification
	IPLimiter       IPLimiter
	IDLimiter       IDLimiter
	ChallengeWindow int64 `env:"CHALLENGE_WINDOW" env-default:"8"`
	Log             Log
	Encryption      Encryption
}

type MongoDB struct {
	URI          string `env:"MONGO_URI" env-default:"mongodb://localhost:27017"`
	DatabaseName string `env:"DATABASE_NAME" env-default:"tfgrid-kyc-db"`
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
}
type TFChain struct {
	WsProviderURL string `env:"TFCHAIN_WS_PROVIDER_URL" env-default:"wss://tfchain.grid.tf"`
}
type Verification struct {
	SuspiciousVerificationOutcome string `env:"VERIFICATION_SUSPICIOUS_VERIFICATION_OUTCOME" env-default:"verified"`
	ExpiredDocumentOutcome        string `env:"VERIFICATION_EXPIRED_DOCUMENT_OUTCOME" env-default:"unverified"`
	MinBalanceToVerifyAccount     uint64 `env:"VERIFICATION_MIN_BALANCE_TO_VERIFY_ACCOUNT" env-default:"10000000"`
}
type IPLimiter struct {
	MaxTokenRequests int `env:"IP_LIMITER_MAX_TOKEN_REQUESTS" env-default:"4"`
	TokenExpiration  int `env:"IP_LIMITER_TOKEN_EXPIRATION" env-default:"24"`
}
type IDLimiter struct {
	MaxTokenRequests int `env:"ID_LIMITER_MAX_TOKEN_REQUESTS" env-default:"4"`
	TokenExpiration  int `env:"ID_LIMITER_TOKEN_EXPIRATION" env-default:"24"`
}
type Log struct {
	Debug bool `env:"DEBUG" env-default:"false"`
}

type Encryption struct {
	Key string `env:"ENCRYPTION_KEY" env-required:"true"`
}

func LoadConfig() (*Config, error) {
	cfg := &Config{}
	err := cleanenv.ReadEnv(cfg)
	if err != nil {
		return nil, errors.NewInternalError("error loading config", err)
	}
	return cfg, nil
}
