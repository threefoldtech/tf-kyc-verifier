package repository

import (
	"context"
	"fmt"

	"github.com/threefoldtech/tf-kyc-verifier/internal/models"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type TokenRepository interface {
	SaveToken(ctx context.Context, token *models.Token) error
	GetToken(ctx context.Context, clientID string) (*models.Token, error)
	DeleteToken(ctx context.Context, clientID string, scanRef string) error
}

type VerificationRepository interface {
	SaveVerification(ctx context.Context, verification *models.Verification) error
	GetVerification(ctx context.Context, clientID string) (*models.Verification, error)
}

func NewMongoClient(ctx context.Context, mongoURI string) (*mongo.Client, error) {
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(mongoURI))
	if err != nil {
		return nil, fmt.Errorf("connecting to MongoDB: %w", err)
	}

	err = client.Ping(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("pinging MongoDB: %w", err)
	}

	return client, nil
}
