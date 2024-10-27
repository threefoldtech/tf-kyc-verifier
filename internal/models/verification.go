package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Verification struct {
	ID                    primitive.ObjectID `bson:"_id,omitempty" json:"-"`
	CreatedAt             time.Time          `bson:"createdAt" json:"-"`
	Final                 *bool              `bson:"final" json:"final"`                     // required
	Platform              Platform           `bson:"platform" json:"platform"`               // required
	Status                Status             `bson:"status" json:"status"`                   // required
	Data                  PersonData         `bson:"data" json:"data"`                       // required
	FileUrls              map[string]string  `bson:"fileUrls" json:"fileUrls"`               // required
	IdenfyRef             string             `bson:"idenfyRef" json:"scanRef"`               // required
	ClientID              string             `bson:"clientId" json:"clientId"`               // required
	StartTime             int64              `bson:"startTime" json:"startTime"`             // required
	FinishTime            int64              `bson:"finishTime" json:"finishTime"`           // required
	ClientIP              string             `bson:"clientIp" json:"clientIp"`               // required
	ClientIPCountry       string             `bson:"clientIpCountry" json:"clientIpCountry"` // required
	ClientLocation        string             `bson:"clientLocation" json:"clientLocation"`   // required
	CompanyID             string             `bson:"companyId" json:"companyId"`             // required
	BeneficiaryID         string             `bson:"beneficiaryId" json:"beneficiaryId"`     // required
	RegistryCenterCheck   interface{}        `json:"registryCenterCheck,omitempty"`
	AddressVerification   interface{}        `json:"addressVerification,omitempty"`
	QuestionnaireAnswers  interface{}        `json:"questionnaireAnswers,omitempty"`
	AdditionalSteps       map[string]string  `json:"additionalSteps,omitempty"`
	UtilityData           []string           `json:"utilityData,omitempty"`
	AdditionalStepPdfUrls map[string]string  `json:"additionalStepPdfUrls,omitempty"`
	AML                   []AMLCheck         `bson:"AML" json:"AML,omitempty"`
	LID                   []LID              `bson:"LID" json:"LID,omitempty"`
	ExternalRef           string             `bson:"externalRef" json:"externalRef,omitempty"`
	ManualAddress         string             `bson:"manualAddress" json:"manualAddress,omitempty"`
	ManualAddressMatch    *bool              `bson:"manualAddressMatch" json:"manualAddressMatch,omitempty"`
}

type Platform string

const (
	PlatformPC        Platform = "PC"
	PlatformMobile    Platform = "MOBILE"
	PlatformTablet    Platform = "TABLET"
	PlatformMobileApp Platform = "MOBILE_APP"
	PlatformMobileSDK Platform = "MOBILE_SDK"
	PlatformOther     Platform = "OTHER"
)

type Overall string

const (
	OverallApproved  Overall = "APPROVED"
	OverallDenied    Overall = "DENIED"
	OverallSuspected Overall = "SUSPECTED"
	OverallReviewing Overall = "REVIEWING"
	OverallExpired   Overall = "EXPIRED"
	OverallActive    Overall = "ACTIVE"
	OverallDeleted   Overall = "DELETED"
	OverallArchived  Overall = "ARCHIVED"
)

type Status struct {
	Overall            *Overall          `bson:"overall" json:"overall"`
	SuspicionReasons   []SuspicionReason `bson:"suspicionReasons" json:"suspicionReasons"`
	DenyReasons        []string          `bson:"denyReasons" json:"denyReasons"`
	FraudTags          []string          `bson:"fraudTags" json:"fraudTags"`
	MismatchTags       []string          `bson:"mismatchTags" json:"mismatchTags"`
	AutoFace           string            `bson:"autoFace" json:"autoFace,omitempty"`
	ManualFace         string            `bson:"manualFace" json:"manualFace,omitempty"`
	AutoDocument       string            `bson:"autoDocument" json:"autoDocument,omitempty"`
	ManualDocument     string            `bson:"manualDocument" json:"manualDocument,omitempty"`
	AdditionalSteps    *AdditionalStep   `bson:"additionalSteps" json:"additionalSteps,omitempty"`
	AMLResultClass     string            `bson:"amlResultClass" json:"amlResultClass,omitempty"`
	PEPSStatus         string            `bson:"pepsStatus" json:"pepsStatus,omitempty"`
	SanctionsStatus    string            `bson:"sanctionsStatus" json:"sanctionsStatus,omitempty"`
	AdverseMediaStatus string            `bson:"adverseMediaStatus" json:"adverseMediaStatus,omitempty"`
}

type SuspicionReason string

const (
	SuspicionFaceSuspected       SuspicionReason = "FACE_SUSPECTED"
	SuspicionFaceBlacklisted     SuspicionReason = "FACE_BLACKLISTED"
	SuspicionDocFaceBlacklisted  SuspicionReason = "DOC_FACE_BLACKLISTED"
	SuspicionDocMobilePhoto      SuspicionReason = "DOC_MOBILE_PHOTO"
	SuspicionDevToolsOpened      SuspicionReason = "DEV_TOOLS_OPENED"
	SuspicionDocPrintSpoofed     SuspicionReason = "DOC_PRINT_SPOOFED"
	SuspicionFakePhoto           SuspicionReason = "FAKE_PHOTO"
	SuspicionAMLSuspection       SuspicionReason = "AML_SUSPECTION"
	SuspicionAMLFailed           SuspicionReason = "AML_FAILED"
	SuspicionLIDSuspection       SuspicionReason = "LID_SUSPECTION"
	SuspicionLIDFailed           SuspicionReason = "LID_FAILED"
	SuspicionSanctionsSuspection SuspicionReason = "SANCTIONS_SUSPECTION"
	SuspicionSanctionsFailed     SuspicionReason = "SANCTIONS_FAILED"
	SuspicionRCFailed            SuspicionReason = "RC_FAILED"
	SuspicionAutoUnverifiable    SuspicionReason = "AUTO_UNVERIFIABLE"
)

type AdditionalStep string

const (
	AdditionalStepValid    AdditionalStep = "VALID"
	AdditionalStepInvalid  AdditionalStep = "INVALID"
	AdditionalStepNotFound AdditionalStep = "NOT_FOUND"
)

type DocumentType string

const (
	ID_CARD                    DocumentType = "ID_CARD"
	PASSPORT                   DocumentType = "PASSPORT"
	RESIDENCE_PERMIT           DocumentType = "RESIDENCE_PERMIT"
	DRIVER_LICENSE             DocumentType = "DRIVER_LICENSE"
	PAN_CARD                   DocumentType = "PAN_CARD"
	AADHAAR                    DocumentType = "AADHAAR"
	OTHER                      DocumentType = "OTHER"
	VISA                       DocumentType = "VISA"
	BORDER_CROSSING            DocumentType = "BORDER_CROSSING"
	ASYLUM                     DocumentType = "ASYLUM"
	NATIONAL_PASSPORT          DocumentType = "NATIONAL_PASSPORT"
	PROVISIONAL_DRIVER_LICENSE DocumentType = "PROVISIONAL_DRIVER_LICENSE"
	VOTER_CARD                 DocumentType = "VOTER_CARD"
	OLD_ID_CARD                DocumentType = "OLD_ID_CARD"
	TRAVEL_CARD                DocumentType = "TRAVEL_CARD"
	PHOTO_CARD                 DocumentType = "PHOTO_CARD"
	MILITARY_CARD              DocumentType = "MILITARY_CARD"
	PROOF_OF_AGE_CARD          DocumentType = "PROOF_OF_AGE_CARD"
	DIPLOMATIC_ID              DocumentType = "DIPLOMATIC_ID"
)

type Sex string

const (
	MALE      Sex = "MALE"
	FEMALE    Sex = "FEMALE"
	UNDEFINED Sex = "UNDEFINED"
)

type AgeEstimate string

const (
	UNDER_13 AgeEstimate = "UNDER_13"
	OVER_13  AgeEstimate = "OVER_13"
	OVER_18  AgeEstimate = "OVER_18"
	OVER_22  AgeEstimate = "OVER_22"
	OVER_25  AgeEstimate = "OVER_25"
	OVER_30  AgeEstimate = "OVER_30"
)

type PersonData struct {
	DocFirstName           string        `bson:"docFirstName" json:"docFirstName"`
	DocLastName            string        `bson:"docLastName" json:"docLastName"`
	DocNumber              string        `bson:"docNumber" json:"docNumber"`
	DocPersonalCode        string        `bson:"docPersonalCode" json:"docPersonalCode"`
	DocExpiry              string        `bson:"docExpiry" json:"docExpiry"`
	DocDOB                 string        `bson:"docDob" json:"docDob"`
	DocDateOfIssue         string        `bson:"docDateOfIssue" json:"docDateOfIssue"`
	DocType                *DocumentType `bson:"docType" json:"docType"`
	DocSex                 *Sex          `bson:"docSex" json:"docSex"`
	DocNationality         string        `bson:"docNationality" json:"docNationality"`
	DocIssuingCountry      string        `bson:"docIssuingCountry" json:"docIssuingCountry"`
	BirthPlace             string        `bson:"birthPlace" json:"birthPlace"`
	Authority              string        `bson:"authority" json:"authority"`
	Address                string        `bson:"address" json:"address"`
	DocTemporaryAddress    string        `bson:"docTemporaryAddress" json:"docTemporaryAddress"`
	MothersMaidenName      string        `bson:"mothersMaidenName" json:"mothersMaidenName"`
	DocBirthName           string        `bson:"docBirthName" json:"docBirthName"`
	DriverLicenseCategory  string        `bson:"driverLicenseCategory" json:"driverLicenseCategory"`
	ManuallyDataChanged    *bool         `bson:"manuallyDataChanged" json:"manuallyDataChanged"`
	FullName               string        `bson:"fullName" json:"fullName"`
	SelectedCountry        string        `bson:"selectedCountry" json:"selectedCountry"`
	OrgFirstName           string        `bson:"orgFirstName" json:"orgFirstName"`
	OrgLastName            string        `bson:"orgLastName" json:"orgLastName"`
	OrgNationality         string        `bson:"orgNationality" json:"orgNationality"`
	OrgBirthPlace          string        `bson:"orgBirthPlace" json:"orgBirthPlace"`
	OrgAuthority           string        `bson:"orgAuthority" json:"orgAuthority"`
	OrgAddress             string        `bson:"orgAddress" json:"orgAddress"`
	OrgTemporaryAddress    string        `bson:"orgTemporaryAddress" json:"orgTemporaryAddress"`
	OrgMothersMaidenName   string        `bson:"orgMothersMaidenName" json:"orgMothersMaidenName"`
	OrgBirthName           string        `bson:"orgBirthName" json:"orgBirthName"`
	AgeEstimate            *AgeEstimate  `bson:"ageEstimate" json:"ageEstimate"`
	ClientIPProxyRiskLevel string        `bson:"clientIpProxyRiskLevel" json:"clientIpProxyRiskLevel"`
	DuplicateFaces         []string      `bson:"duplicateFaces" json:"duplicateFaces"`
	DuplicateDocFaces      []string      `bson:"duplicateDocFaces" json:"duplicateDocFaces"`
	AdditionalData         interface{}   `bson:"additionalData" json:"additionalData"`
}

type AMLCheck struct {
	Status           ServiceStatus `bson:"status"`
	Data             []AMLData     `bson:"data"`
	ServiceName      string        `bson:"serviceName"`
	ServiceGroupType string        `bson:"serviceGroupType"`
	UID              string        `bson:"uid"`
	ErrorMessage     string        `bson:"errorMessage"`
}

type AMLData struct {
	Name        string   `bson:"name"`
	Surname     string   `bson:"surname"`
	Nationality string   `bson:"nationality"`
	DOB         string   `bson:"dob"`
	Suspicion   string   `bson:"suspicion"`
	Reason      string   `bson:"reason"`
	ListNumber  string   `bson:"listNumber"`
	ListName    string   `bson:"listName"`
	Score       *float64 `bson:"score"`
	LastUpdate  *string  `bson:"lastUpdate"`
	IsPerson    *bool    `bson:"isPerson"`
	IsActive    *bool    `bson:"isActive"`
	CheckDate   string   `bson:"checkDate"`
}

type LID struct {
	Status           *ServiceStatus `json:"status"`
	Data             []LIDData      `json:"data"`
	ServiceName      string         `json:"serviceName"`
	ServiceGroupType string         `json:"serviceGroupType"`
	UID              string         `json:"uid"`
	ErrorMessage     string         `json:"errorMessage"`
}

type LIDData struct {
	DocumentNumber string        `json:"documentNumber"`
	DocumentType   *DocumentType `json:"documentType"`
	Valid          *bool         `json:"valid"`
	ExpiryDate     string        `json:"expiryDate"`
	CheckDate      string        `json:"checkDate"`
}

type ServiceStatus struct {
	ServiceSuspected *bool  `json:"serviceSuspected" bson:"serviceSuspected"`
	CheckSuccessful  *bool  `json:"checkSuccessful" bson:"checkSuccessful"`
	ServiceFound     *bool  `json:"serviceFound" bson:"serviceFound"`
	ServiceUsed      *bool  `json:"serviceUsed" bson:"serviceUsed"`
	OverallStatus    string `json:"overallStatus" bson:"overallStatus"`
}

type VerificationOutcome struct {
	Final     *bool   `bson:"final"`
	ClientID  string  `bson:"clientId"`
	IdenfyRef string  `bson:"idenfyRef"`
	Outcome   Outcome `bson:"outcome"`
}

type Outcome string

const (
	OutcomeApproved Outcome = "APPROVED"
	OutcomeRejected Outcome = "REJECTED"
)
