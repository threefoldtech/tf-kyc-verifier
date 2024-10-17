package responses

import (
	"example.com/tfgrid-kyc-service/internal/models"
	"github.com/gofiber/fiber/v2"
)

type Response struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Message string      `json:"message,omitempty"`
}

func SuccessResponse(c *fiber.Ctx, statusCode int, data interface{}, message string) error {
	return c.Status(statusCode).JSON(Response{
		Success: true,
		Data:    data,
		Message: message,
	})
}

type TokenResponseWithStatus struct {
	Token      *TokenResponse `json:"token"`
	IsNewToken bool           `json:"is_new_token"`
	Message    string         `json:"message"`
}

type TokenResponse struct {
	AuthToken     string `json:"authToken"`
	ScanRef       string `json:"scanRef"`
	ClientID      string `json:"clientId"`
	ExpiryTime    int    `json:"expiryTime"`
	SessionLength int    `json:"sessionLength"`
	DigitString   string `json:"digitString"`
	TokenType     string `json:"tokenType"`
}

type VerificationStatusResponse struct {
	FraudTags      []string `json:"fraudTags"`
	MismatchTags   []string `json:"mismatchTags"`
	AutoDocument   string   `json:"autoDocument"`
	AutoFace       string   `json:"autoFace"`
	ManualDocument string   `json:"manualDocument"`
	ManualFace     string   `json:"manualFace"`
	ScanRef        string   `json:"scanRef"`
	ClientID       string   `json:"clientId"`
	Status         string   `json:"status"`
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
	ManuallyDataChanged    bool        `json:"manuallyDataChanged"`
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
	ScanRef                string      `json:"scanRef"`
	ClientID               string      `json:"clientId"`
}

// implement from() method for TokenResponseWithStatus
func NewTokenResponseWithStatus(token *models.Token, isNewToken bool) *TokenResponseWithStatus {
	message := "Existing valid token retrieved."
	if isNewToken {
		message = "New token created."
	}
	return &TokenResponseWithStatus{
		Token: &TokenResponse{
			AuthToken:     token.AuthToken,
			ScanRef:       token.ScanRef,
			ClientID:      token.ClientID,
			ExpiryTime:    token.ExpiryTime,
			SessionLength: token.SessionLength,
			DigitString:   token.DigitString,
			TokenType:     token.TokenType,
		},
		IsNewToken: isNewToken,
		Message:    message,
	}
}

func NewVerificationStatusResponse(verification *models.Verification) *VerificationStatusResponse {
	return &VerificationStatusResponse{
		FraudTags:      verification.Status.FraudTags,
		MismatchTags:   verification.Status.MismatchTags,
		AutoDocument:   verification.Status.AutoDocument,
		ManualDocument: verification.Status.ManualDocument,
		AutoFace:       verification.Status.AutoFace,
		ManualFace:     verification.Status.ManualFace,
		ScanRef:        verification.ScanRef,
		ClientID:       verification.ClientID,
		Status:         string(verification.Status.Overall),
	}
}

func NewVerificationDataResponse(verification *models.Verification) *VerificationDataResponse {
	return &VerificationDataResponse{
		DocFirstName:           verification.Data.DocFirstName,
		DocLastName:            verification.Data.DocLastName,
		DocNumber:              verification.Data.DocNumber,
		DocPersonalCode:        verification.Data.DocPersonalCode,
		DocExpiry:              verification.Data.DocExpiry,
		DocDob:                 verification.Data.DocDOB,
		DocDateOfIssue:         verification.Data.DocDateOfIssue,
		DocType:                string(verification.Data.DocType),
		DocSex:                 string(verification.Data.DocSex),
		DocNationality:         verification.Data.DocNationality,
		DocIssuingCountry:      verification.Data.DocIssuingCountry,
		DocTemporaryAddress:    verification.Data.DocTemporaryAddress,
		DocBirthName:           verification.Data.DocBirthName,
		BirthPlace:             verification.Data.BirthPlace,
		Authority:              verification.Data.Authority,
		MotherMaidenName:       verification.Data.MothersMaidenName,
		DriverLicenseCategory:  verification.Data.DriverLicenseCategory,
		ManuallyDataChanged:    verification.Data.ManuallyDataChanged,
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
		AgeEstimate:            string(verification.Data.AgeEstimate),
		ClientIpProxyRiskLevel: verification.Data.ClientIPProxyRiskLevel,
		DuplicateFaces:         verification.Data.DuplicateFaces,
		DuplicateDocFaces:      verification.Data.DuplicateDocFaces,
		AddressVerification:    verification.AddressVerification,
		AdditionalData:         verification.Data.AdditionalData,
		ScanRef:                verification.ScanRef,
		ClientID:               verification.ClientID,
	}
}
