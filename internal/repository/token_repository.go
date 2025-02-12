package repository

import (
	"context"
	"time"

	"log/slog"

	"github.com/threefoldtech/tf-kyc-verifier/internal/metrics"
	"github.com/threefoldtech/tf-kyc-verifier/internal/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type MongoTokenRepository struct {
	collection *mongo.Collection
	logger     *slog.Logger
	metrics    *metrics.Metrics
}

func NewMongoTokenRepository(ctx context.Context, db *mongo.Database, logger *slog.Logger) TokenRepository {
	repo := &MongoTokenRepository{
		collection: db.Collection("tokens"),
		logger:     logger,
		metrics:    metrics.GetInstance(),
	}
	repo.createTTLIndex(ctx)
	repo.createCollectionIndexes(ctx)
	return repo
}

func (r *MongoTokenRepository) createTTLIndex(ctx context.Context) {
	start := time.Now()
	_, err := r.collection.Indexes().CreateOne(
		ctx,
		mongo.IndexModel{
			Keys:    bson.D{{Key: "expiresAt", Value: 1}},
			Options: options.Index().SetExpireAfterSeconds(0),
		},
	)
	r.metrics.MongoDBOperationsLatency.WithLabelValues("create_ttl_index", "token").Observe(time.Since(start).Seconds())
	if err != nil {
		r.logger.Error("Error creating TTL index", "error", err)
		r.metrics.MongoDBOperationsError.WithLabelValues("create_ttl_index", "token").Inc()
	}
}

func (r *MongoTokenRepository) createCollectionIndexes(ctx context.Context) {
	keys := []bson.D{
		{{Key: "clientId", Value: 1}},
		{{Key: "scanRef", Value: 1}},
	}
	for _, key := range keys {
		start := time.Now()
		_, err := r.collection.Indexes().CreateOne(ctx, mongo.IndexModel{
			Keys:    key,
			Options: options.Index().SetUnique(true),
		})
		r.metrics.MongoDBOperationsLatency.WithLabelValues("create_index", "token").Observe(time.Since(start).Seconds())
		if err != nil {
			r.logger.Error("Error creating index", "key", key, "error", err)
			r.metrics.MongoDBOperationsError.WithLabelValues("create_index", "token").Inc()
		}
	}
}

func (r *MongoTokenRepository) SaveToken(ctx context.Context, token *models.Token) error {
	token.CreatedAt = time.Now()
	token.ExpiresAt = token.CreatedAt.Add(time.Duration(token.ExpiryTime) * time.Second)
	_, err := r.collection.InsertOne(ctx, token)
	r.metrics.MongoDBOperationsLatency.WithLabelValues("insert", "token").Observe(time.Since(token.CreatedAt).Seconds())
	if err != nil {
		r.logger.Error("Error saving token", "error", err)
		r.metrics.MongoDBOperationsError.WithLabelValues("insert", "token").Inc()
	}
	return err
}

func (r *MongoTokenRepository) GetToken(ctx context.Context, clientID string) (*models.Token, error) {
	var token models.Token
	start := time.Now()
	err := r.collection.FindOne(ctx, bson.M{"clientId": clientID}).Decode(&token)
	r.metrics.MongoDBOperationsLatency.WithLabelValues("find_one", "token").Observe(time.Since(start).Seconds())
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		r.metrics.MongoDBOperationsError.WithLabelValues("fine_one", "token").Inc()
		return nil, err
	}
	return &token, nil
}

func (r *MongoTokenRepository) DeleteToken(ctx context.Context, clientID string, scanRef string) error {
	start := time.Now()
	_, err := r.collection.DeleteOne(ctx, bson.M{"clientId": clientID, "scanRef": scanRef})
	r.metrics.MongoDBOperationsLatency.WithLabelValues("delete", "token").Observe(time.Since(start).Seconds())
	if err != nil {
		r.logger.Error("Error deleting token", "error", err)
		r.metrics.MongoDBOperationsError.WithLabelValues("delete", "token").Inc()
	}
	return err
}
