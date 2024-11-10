package models

type ExpirationThreshold string

const (
	ExpiresWithin30Days ExpirationThreshold = "DOCUMENT_EXPIRES_WITHIN_30_DAYS"
	ExpiresWithin7Days  ExpirationThreshold = "DOCUMENT_EXPIRES_WITHIN_7_DAYS"
	ExpiresWithin1Day   ExpirationThreshold = "DOCUMENT_EXPIRES_WITHIN_1_DAY"
	DocumentExpired     ExpirationThreshold = "DOCUMENT_EXPIRED"
)

type DocExpirationNotification struct {
	ScanRef             string              `json:"scanRef" bson:"scanRef"`
	ClientID            string              `json:"clientId" bson:"clientId"`
	ExpirationThreshold ExpirationThreshold `json:"expirationThreshold" bson:"expirationThreshold"`
	DocumentExpiration  string              `json:"documentExpiration" bson:"documentExpiration"`
}
