package services

import (
	"context"

	"example.com/tfgrid-kyc-service/internal/models"
)

type KYCService interface {
	GetorCreateVerificationToken(ctx context.Context, clientID string) (*models.Token, bool, error)
	DeleteToken(ctx context.Context, clientID string, scanRef string) error
	AccountHasRequiredBalance(ctx context.Context, address string) (bool, error)
	GetVerification(ctx context.Context, clientID string) (*models.Verification, error)
	GetVerificationStatus(ctx context.Context, clientID string) (*models.VerificationOutcome, error)
	GetVerificationStatusByTwinID(ctx context.Context, twinID string) (*models.VerificationOutcome, error)
	ProcessVerificationResult(ctx context.Context, body []byte, sigHeader string, result models.Verification) error
	ProcessDocExpirationNotification(ctx context.Context, clientID string) error
	IsUserVerified(ctx context.Context, clientID string) (bool, error)
}
