package repository

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/threefoldtech/tf-kyc-verifier/internal/metrics"
	"github.com/threefoldtech/tf-kyc-verifier/internal/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type MongoVerificationRepository struct {
	collection *mongo.Collection
	logger     *slog.Logger
	metrics    *metrics.Metrics
}

func NewMongoVerificationRepository(ctx context.Context, db *mongo.Database, logger *slog.Logger) VerificationRepository {
	// create index for clientId
	repo := &MongoVerificationRepository{
		collection: db.Collection("verifications"),
		logger:     logger,
		metrics:    metrics.GetInstance(),
	}
	repo.createCollectionIndexes(ctx)
	return repo
}

func (r *MongoVerificationRepository) createCollectionIndexes(ctx context.Context) {
	key := bson.D{{Key: "clientId", Value: 1}}
	start := time.Now()
	_, err := r.collection.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    key,
		Options: options.Index().SetUnique(false),
	})
	r.metrics.MongoDBOperationsLatency.WithLabelValues("create_index","verification").Observe(time.Since(start).Seconds())
	if err != nil {
		r.logger.Error("Error creating index", "key", key, "error", err)
		r.metrics.MongoDBOperationsError.WithLabelValues("create_index","verification").Inc()
	}
}

func (r *MongoVerificationRepository) SaveVerification(ctx context.Context, verification *models.Verification) error {
	verification.CreatedAt = time.Now()
	_, err := r.collection.InsertOne(ctx, verification)
	r.metrics.MongoDBOperationsLatency.WithLabelValues("insert","verification").Observe(time.Since(verification.CreatedAt).Seconds())
	if err != nil {
		r.logger.Error("Error saving verification", "error", err)
		r.metrics.MongoDBOperationsError.WithLabelValues("insert","verification").Inc()
	}
	return err
}

func (r *MongoVerificationRepository) GetVerification(ctx context.Context, clientID string) (*models.Verification, error) {
	var verification models.Verification
	// return the latest verification
	opts := options.FindOne().SetSort(bson.D{{Key: "createdAt", Value: -1}})
	start := time.Now()
	err := r.collection.FindOne(ctx, bson.M{"clientId": clientID}, opts).Decode(&verification)
	r.metrics.MongoDBOperationsLatency.WithLabelValues("find_one","verification").Observe(time.Since(start).Seconds())
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		r.logger.Error("Error getting verification", "error", err)
		r.metrics.MongoDBOperationsError.WithLabelValues("find_one","verification").Inc()
		return nil, err
	}
	return &verification, nil
}

func (r *MongoVerificationRepository) UpdateExpirationStatus(ctx context.Context, clientID string, scanRef string, status models.ExpirationThreshold) error {
	filter := bson.M{"clientId": clientID, "scanRef": scanRef}
	update := bson.M{
		"$set": bson.M{
			"expirationStatus": status,
		},
	}
	start:= time.Now()
	result, err := r.collection.UpdateOne(ctx, filter, update)
	r.metrics.MongoDBOperationsLatency.WithLabelValues("update","verification").Observe(time.Since(start).Seconds())
	if err != nil {
		r.metrics.MongoDBOperationsError.WithLabelValues("update","verification").Inc()
		return fmt.Errorf("updating expiration status: %w", err)
	}

	if result.MatchedCount == 0 {
		return fmt.Errorf("verification not found for client: %s", clientID)
	}

	return nil
}
