package services

import (
	"context"

	"github.com/threefoldtech/tf-kyc-verifier/internal/models"
)

type KYCService interface {
	GetOrCreateVerificationToken(ctx context.Context, clientID string) (*models.Token, bool, error)
	DeleteToken(ctx context.Context, clientID string, scanRef string) error
	AccountHasRequiredBalance(ctx context.Context, address string) (bool, error)
	GetVerificationData(ctx context.Context, clientID string) (*models.Verification, error)
	GetVerificationStatus(ctx context.Context, clientID string) (*models.VerificationOutcome, error)
	GetVerificationStatusByTwinID(ctx context.Context, twinID string) (*models.VerificationOutcome, error)
	ProcessVerificationResult(ctx context.Context, body []byte, sigHeader string, result models.Verification) error
	ProcessDocExpirationNotification(ctx context.Context, clientID string) error
	IsUserVerified(ctx context.Context, clientID string) (bool, error)
}
