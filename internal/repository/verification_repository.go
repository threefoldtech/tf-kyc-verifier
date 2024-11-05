package repository

import (
	"context"
	"log/slog"
	"time"

	"github.com/threefoldtech/tf-kyc-verifier/internal/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type MongoVerificationRepository struct {
	collection *mongo.Collection
	logger     *slog.Logger
}

func NewMongoVerificationRepository(ctx context.Context, db *mongo.Database, logger *slog.Logger) VerificationRepository {
	// create index for clientId
	repo := &MongoVerificationRepository{
		collection: db.Collection("verifications"),
		logger:     logger,
	}
	repo.createClientIdIndex(ctx)
	return repo
}

func (r *MongoVerificationRepository) createClientIdIndex(ctx context.Context) {
	_, err := r.collection.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.D{{Key: "clientId", Value: 1}},
		Options: options.Index().SetUnique(true),
	})
	if err != nil {
		r.logger.Error("Error creating clientId index", "error", err)
	}
}

func (r *MongoVerificationRepository) SaveVerification(ctx context.Context, verification *models.Verification) error {
	verification.CreatedAt = time.Now()
	_, err := r.collection.InsertOne(ctx, verification)
	return err
}

func (r *MongoVerificationRepository) GetVerification(ctx context.Context, clientID string) (*models.Verification, error) {
	var verification models.Verification
	// return the latest verification
	opts := options.FindOne().SetSort(bson.D{{Key: "createdAt", Value: -1}})
	err := r.collection.FindOne(ctx, bson.M{"clientId": clientID}, opts).Decode(&verification)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}
	return &verification, nil
}
