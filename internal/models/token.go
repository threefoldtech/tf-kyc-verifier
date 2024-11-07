package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Token struct {
	ID               primitive.ObjectID  `bson:"_id,omitempty"`
	AuthToken        string              `bson:"authToken"`
	ScanRef          string              `bson:"scanRef"`
	ClientID         string              `bson:"clientId"`
	FirstName        string              `bson:"firstName"`
	LastName         string              `bson:"lastName"`
	SuccessURL       string              `bson:"successUrl"`
	ErrorURL         string              `bson:"errorUrl"`
	UnverifiedURL    string              `bson:"unverifiedUrl"`
	CallbackURL      string              `bson:"callbackUrl"`
	Locale           string              `bson:"locale"`
	ShowInstructions bool                `bson:"showInstructions"`
	Country          string              `bson:"country"`
	ExpiryTime       int                 `bson:"expiryTime"`
	SessionLength    int                 `bson:"sessionLength"`
	Documents        []string            `bson:"documents"`
	AllowedDocuments map[string][]string `bson:"allowedDocuments"`
	DateOfBirth      string              `bson:"dateOfBirth"`
	DateOfExpiry     string              `bson:"dateOfExpiry"`
	DateOfIssue      string              `bson:"dateOfIssue"`
	Nationality      string              `bson:"nationality"`
	PersonalNumber   string              `bson:"personalNumber"`
	DocumentNumber   string              `bson:"documentNumber"`
	Sex              string              `bson:"sex"`
	DigitString      string              `bson:"digitString"`
	Address          string              `bson:"address"`
	TokenType        string              `bson:"tokenType"`
	ExternalRef      string              `bson:"externalRef"`
	Questionnaire    interface{}         `bson:"questionnaire"`
	UtilityBill      bool                `bson:"utilityBill"`
	AdditionalSteps  interface{}         `bson:"additionalSteps"`
	AdditionalData   interface{}         `bson:"additionalData"`
	CreatedAt        time.Time           `bson:"createdAt"`
	ExpiresAt        time.Time           `bson:"expiresAt"`
}
