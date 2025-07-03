package repository

import (
	"context"
	"log/slog"
	"time"

	"github.com/threefoldtech/tf-kyc-verifier/internal/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// PaginationParams represents pagination parameters for listing sponsorships
type PaginationParams struct {
	Limit  int64
	Offset int64
}

type SponsorshipRepository interface {
	// Create creates a new sponsorship
	// Returns mongo.WriteException with code 11000 if a sponsorship already exists for the sponsee
	Create(ctx context.Context, sponsorship *models.Sponsorship) error

	// GetBySponsor returns all active sponsorships for a given sponsor
	GetBySponsor(ctx context.Context, sponsorClientID string, pagination PaginationParams) ([]*models.Sponsorship, int64, error)

	// GetBySponsee returns the active sponsorship for a given sponsee
	// Returns nil, nil if no active sponsorship exists
	GetBySponsee(ctx context.Context, sponseeClientID string) (*models.Sponsorship, error)

	// ListAll returns a paginated list of all active sponsorships
	// Returns the list of sponsorships and the total count
	ListAll(ctx context.Context, pagination PaginationParams) ([]*models.Sponsorship, int64, error)
}

type mongoSponsorshipRepository struct {
	collection *mongo.Collection
	logger     *slog.Logger
}

// NewMongoSponsorshipRepository creates a new MongoDB implementation of SponsorshipRepository
func NewMongoSponsorshipRepository(ctx context.Context, db *mongo.Database, logger *slog.Logger) SponsorshipRepository {
	repo := &mongoSponsorshipRepository{
		collection: db.Collection((&models.Sponsorship{}).CollectionName()),
		logger:     logger,
	}

	// Create indexes
	repo.createCollectionIndexes(ctx)

	return repo
}

// Create a new sponsorship
// Returns mongo.WriteException with code 11000 if a sponsorship already exists for the sponsee
func (r *mongoSponsorshipRepository) Create(ctx context.Context, sponsorship *models.Sponsorship) error {
	sponsorship.CreatedAt = time.Now()
	

	// This will fail with a duplicate key error if a sponsorship already exists for this sponsee
	// due to the unique index on sponsee_twin_id with is_active: true
	result, err := r.collection.InsertOne(ctx, sponsorship)
	if err != nil {
		return err
	}

	if oid, ok := result.InsertedID.(primitive.ObjectID); ok {
		sponsorship.ID = oid
	}

	return nil
}

func (r *mongoSponsorshipRepository) GetBySponsor(ctx context.Context, sponsorClientID string, pagination PaginationParams) ([]*models.Sponsorship, int64, error) {
	filter := bson.M{
		"sponsor_client_id": sponsorClientID,
	}

	// Count total documents first
	total, err := r.collection.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, err
	}

	// Create find options for pagination
	findOptions := options.Find()
	findOptions.SetLimit(pagination.Limit)
	findOptions.SetSkip(pagination.Offset)
	findOptions.SetSort(bson.D{{Key: "created_at", Value: -1}}) // Sort by creation date, newest first

	cursor, err := r.collection.Find(ctx, filter, findOptions)
	if err != nil {
		return nil, 0, err
	}
	defer cursor.Close(ctx)

	var sponsorships []*models.Sponsorship
	if err := cursor.All(ctx, &sponsorships); err != nil {
		return nil, 0, err
	}

	return sponsorships, total, nil
}

func (r *mongoSponsorshipRepository) GetBySponsee(ctx context.Context, sponseeClientID string) (*models.Sponsorship, error) {
	var sponsorship models.Sponsorship
	err := r.collection.FindOne(ctx, bson.M{
		"sponsee_client_id": sponseeClientID,
	}).Decode(&sponsorship)

	if err == mongo.ErrNoDocuments {
		return nil, nil
	}

	if err != nil {
		return nil, err
	}

	return &sponsorship, nil
}

func (r *mongoSponsorshipRepository) ListAll(ctx context.Context, pagination PaginationParams) ([]*models.Sponsorship, int64, error) {
	// Count total documents first
	total, err := r.collection.CountDocuments(ctx, bson.M{})
	if err != nil {
		return nil, 0, err
	}

	// Create find options for pagination
	findOptions := options.Find()
	findOptions.SetLimit(pagination.Limit)
	findOptions.SetSkip(pagination.Offset)
	findOptions.SetSort(bson.D{{Key: "created_at", Value: -1}}) // Sort by creation date, newest first

	// Find all active sponsorships with pagination
	cursor, err := r.collection.Find(ctx, bson.M{}, findOptions)
	if err != nil {
		return nil, 0, err
	}
	defer cursor.Close(ctx)

	var sponsorships []*models.Sponsorship
	if err := cursor.All(ctx, &sponsorships); err != nil {
		return nil, 0, err
	}

	return sponsorships, total, nil
}

func (r *mongoSponsorshipRepository) createCollectionIndexes(ctx context.Context) {
	// Index for sponsor lookups, optimized for sorting
	sponsorIndex := mongo.IndexModel{
		Keys: bson.D{
			{Key: "sponsor_client_id", Value: 1},
			{Key: "created_at", Value: -1},
		},
	}

	// Unique index for sponsee (a twin can only be sponsored once)
	sponseeIndex := mongo.IndexModel{
		Keys:    bson.D{{Key: "sponsee_client_id", Value: 1}},
		Options: options.Index().SetUnique(true),
	}

	// Index for listing all sponsorships, sorted by creation date
	listAllIndex := mongo.IndexModel{
		Keys: bson.D{{Key: "created_at", Value: -1}},
	}

	_, err := r.collection.Indexes().CreateMany(ctx, []mongo.IndexModel{sponsorIndex, sponseeIndex, listAllIndex})
	if err != nil {
		r.logger.Error("Error creating sponsorship indexes", "error", err)
	}
}
