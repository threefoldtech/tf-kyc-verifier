package repository

import (
	"context"
	"time"

	"example.com/tfgrid-kyc-service/internal/logger"
	"example.com/tfgrid-kyc-service/internal/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type MongoVerificationRepository struct {
	collection *mongo.Collection
	logger     *logger.Logger
}

func NewMongoVerificationRepository(db *mongo.Database, logger *logger.Logger) VerificationRepository {
	return &MongoVerificationRepository{
		collection: db.Collection("verifications"),
		logger:     logger,
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
