package repository

import (
	"context"
	"time"

	"log/slog"

	"github.com/threefoldtech/tf-kyc-verifier/internal/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type MongoTokenRepository struct {
	collection *mongo.Collection
	logger     *slog.Logger
}

func NewMongoTokenRepository(ctx context.Context, db *mongo.Database, logger *slog.Logger) TokenRepository {
	repo := &MongoTokenRepository{
		collection: db.Collection("tokens"),
		logger:     logger,
	}
	repo.createTTLIndex(ctx)
	repo.createClientIdIndex(ctx)
	return repo
}

func (r *MongoTokenRepository) createTTLIndex(ctx context.Context) {
	_, err := r.collection.Indexes().CreateOne(
		ctx,
		mongo.IndexModel{
			Keys:    bson.D{{Key: "expiresAt", Value: 1}},
			Options: options.Index().SetExpireAfterSeconds(0),
		},
	)
	if err != nil {
		r.logger.Error("Error creating TTL index", "error", err)
	}
}

func (r *MongoTokenRepository) createClientIdIndex(ctx context.Context) {
	_, err := r.collection.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.D{{Key: "clientId", Value: 1}},
		Options: options.Index().SetUnique(true),
	})
	if err != nil {
		r.logger.Error("Error creating clientId index", "error", err)
	}
}

func (r *MongoTokenRepository) SaveToken(ctx context.Context, token *models.Token) error {
	token.CreatedAt = time.Now()
	token.ExpiresAt = token.CreatedAt.Add(time.Duration(token.ExpiryTime) * time.Second)
	_, err := r.collection.InsertOne(ctx, token)
	return err
}

func (r *MongoTokenRepository) GetToken(ctx context.Context, clientID string) (*models.Token, error) {
	var token models.Token
	err := r.collection.FindOne(ctx, bson.M{"clientId": clientID}).Decode(&token)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}

	return &token, nil
}

func (r *MongoTokenRepository) DeleteToken(ctx context.Context, clientID string, scanRef string) error {
	_, err := r.collection.DeleteOne(ctx, bson.M{"clientId": clientID, "scanRef": scanRef})
	return err
}
