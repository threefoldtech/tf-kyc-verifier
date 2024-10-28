package responses

import (
	"example.com/tfgrid-kyc-service/internal/models"
)

type ErrorResponse struct {
	Error string `json:"error"`
}

type HealthStatus string

const (
	HealthStatusHealthy  HealthStatus = "Healthy"
	HealthStatusDegraded HealthStatus = "Degraded"
)

type HealthResponse struct {
	Status    HealthStatus `json:"status"`
	Timestamp string       `json:"timestamp"`
	Errors    []string     `json:"errors"`
}

type TokenResponse struct {
	Message       string `json:"message"`
	AuthToken     string `json:"authToken"`
	ScanRef       string `json:"scanRef"`
	ClientID      string `json:"clientId"`
	ExpiryTime    int    `json:"expiryTime"`
	SessionLength int    `json:"sessionLength"`
	DigitString   string `json:"digitString"`
	TokenType     string `json:"tokenType"`
}

type Outcome string

const (
	OutcomeVerified Outcome = "VERIFIED"
	OutcomeRejected Outcome = "REJECTED"
)

type VerificationStatusResponse struct {
	Final     bool    `json:"final"`
	IdenfyRef string  `json:"idenfyRef"`
	ClientID  string  `json:"clientId"`
	Status    Outcome `json:"status"`
}

type VerificationDataResponse struct {
	DocFirstName           string      `json:"docFirstName"`
	DocLastName            string      `json:"docLastName"`
	DocNumber              string      `json:"docNumber"`
	DocPersonalCode        string      `json:"docPersonalCode"`
	DocExpiry              string      `json:"docExpiry"`
	DocDob                 string      `json:"docDob"`
	DocDateOfIssue         string      `json:"docDateOfIssue"`
	DocType                string      `json:"docType"`
	DocSex                 string      `json:"docSex"`
	DocNationality         string      `json:"docNationality"`
	DocIssuingCountry      string      `json:"docIssuingCountry"`
	DocTemporaryAddress    string      `json:"docTemporaryAddress"`
	DocBirthName           string      `json:"docBirthName"`
	BirthPlace             string      `json:"birthPlace"`
	Authority              string      `json:"authority"`
	Address                string      `json:"address"`
	MotherMaidenName       string      `json:"mothersMaidenName"`
	DriverLicenseCategory  string      `json:"driverLicenseCategory"`
	ManuallyDataChanged    *bool       `json:"manuallyDataChanged"`
	FullName               string      `json:"fullName"`
	OrgFirstName           string      `json:"orgFirstName"`
	OrgLastName            string      `json:"orgLastName"`
	OrgNationality         string      `json:"orgNationality"`
	OrgBirthPlace          string      `json:"orgBirthPlace"`
	OrgAuthority           string      `json:"orgAuthority"`
	OrgAddress             string      `json:"orgAddress"`
	OrgTemporaryAddress    string      `json:"orgTemporaryAddress"`
	OrgMothersMaidenName   string      `json:"orgMothersMaidenName"`
	OrgBirthName           string      `json:"orgBirthName"`
	SelectedCountry        string      `json:"selectedCountry"`
	AgeEstimate            string      `json:"ageEstimate"`
	ClientIpProxyRiskLevel string      `json:"clientIpProxyRiskLevel"`
	DuplicateFaces         []string    `json:"duplicateFaces"`
	DuplicateDocFaces      []string    `json:"duplicateDocFaces"`
	AddressVerification    interface{} `json:"addressVerification"`
	AdditionalData         interface{} `json:"additionalData"`
	IdenfyRef              string      `json:"idenfyRef"`
	ClientID               string      `json:"clientId"`
}

func NewTokenResponseWithStatus(token *models.Token, isNewToken bool) *TokenResponse {
	message := "Existing valid token retrieved."
	if isNewToken {
		message = "New token created."
	}
	return &TokenResponse{
		AuthToken:     token.AuthToken,
		ScanRef:       token.ScanRef,
		ClientID:      token.ClientID,
		ExpiryTime:    token.ExpiryTime,
		SessionLength: token.SessionLength,
		DigitString:   token.DigitString,
		TokenType:     token.TokenType,
		Message:       message,
	}
}

func NewVerificationStatusResponse(verificationOutcome *models.VerificationOutcome) *VerificationStatusResponse {
	outcome := OutcomeVerified
	if verificationOutcome.Outcome == "REJECTED" {
		outcome = OutcomeRejected
	}
	return &VerificationStatusResponse{
		Final:     *verificationOutcome.Final,
		IdenfyRef: verificationOutcome.IdenfyRef,
		ClientID:  verificationOutcome.ClientID,
		Status:    outcome,
	}
}

func NewVerificationDataResponse(verification *models.Verification) *VerificationDataResponse {
	var docType string
	if verification.Data.DocType != nil {
		docType = string(*verification.Data.DocType)
	}
	var docSex string
	if verification.Data.DocSex != nil {
		docSex = string(*verification.Data.DocSex)
	}
	var manuallyDataChanged *bool
	if verification.Data.ManuallyDataChanged != nil {
		manuallyDataChanged = verification.Data.ManuallyDataChanged
	}
	var ageEstimate string
	if verification.Data.AgeEstimate != nil {
		ageEstimate = string(*verification.Data.AgeEstimate)
	}
	return &VerificationDataResponse{
		DocFirstName:           verification.Data.DocFirstName,
		DocLastName:            verification.Data.DocLastName,
		DocNumber:              verification.Data.DocNumber,
		DocPersonalCode:        verification.Data.DocPersonalCode,
		DocExpiry:              verification.Data.DocExpiry,
		DocDob:                 verification.Data.DocDOB,
		DocDateOfIssue:         verification.Data.DocDateOfIssue,
		DocType:                docType,
		DocSex:                 docSex,
		DocNationality:         verification.Data.DocNationality,
		DocIssuingCountry:      verification.Data.DocIssuingCountry,
		DocTemporaryAddress:    verification.Data.DocTemporaryAddress,
		DocBirthName:           verification.Data.DocBirthName,
		BirthPlace:             verification.Data.BirthPlace,
		Authority:              verification.Data.Authority,
		MotherMaidenName:       verification.Data.MothersMaidenName,
		DriverLicenseCategory:  verification.Data.DriverLicenseCategory,
		ManuallyDataChanged:    manuallyDataChanged,
		FullName:               verification.Data.FullName,
		OrgFirstName:           verification.Data.OrgFirstName,
		OrgLastName:            verification.Data.OrgLastName,
		OrgNationality:         verification.Data.OrgNationality,
		OrgBirthPlace:          verification.Data.OrgBirthPlace,
		OrgAuthority:           verification.Data.OrgAuthority,
		OrgAddress:             verification.Data.OrgAddress,
		OrgTemporaryAddress:    verification.Data.OrgTemporaryAddress,
		OrgMothersMaidenName:   verification.Data.OrgMothersMaidenName,
		OrgBirthName:           verification.Data.OrgBirthName,
		SelectedCountry:        verification.Data.SelectedCountry,
		AgeEstimate:            ageEstimate,
		ClientIpProxyRiskLevel: verification.Data.ClientIPProxyRiskLevel,
		DuplicateFaces:         verification.Data.DuplicateFaces,
		DuplicateDocFaces:      verification.Data.DuplicateDocFaces,
		AddressVerification:    verification.AddressVerification,
		AdditionalData:         verification.Data.AdditionalData,
		IdenfyRef:              verification.IdenfyRef,
		ClientID:               verification.ClientID,
	}
}
