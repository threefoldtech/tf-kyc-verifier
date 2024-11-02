package idenfy

import (
	"context"

	"example.com/tfgrid-kyc-service/internal/models"
)

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

type IdenfyClient interface {
	CreateVerificationSession(ctx context.Context, clientID string) (models.Token, error)
	VerifyCallbackSignature(ctx context.Context, body []byte, sigHeader string) error
}
