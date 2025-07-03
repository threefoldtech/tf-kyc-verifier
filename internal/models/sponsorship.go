package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Sponsorship represents a relationship where a KYC-verified twin (sponsor)
// sponsors another twin (sponsee) for KYC verification purposes
type Sponsorship struct {
	ID              primitive.ObjectID `bson:"_id,omitempty"`
	SponsorClientID string             `bson:"sponsor_client_id"`
	SponseeClientID string             `bson:"sponsee_client_id"`
	CreatedAt       time.Time          `bson:"created_at"`
	
}

// CollectionName returns the name of the MongoDB collection for sponsorships
func (s *Sponsorship) CollectionName() string {
	return "sponsorships"
}
