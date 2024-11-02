package idenfy

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"testing"

	"example.com/tfgrid-kyc-service/internal/configs"
	"example.com/tfgrid-kyc-service/internal/logger"
	"example.com/tfgrid-kyc-service/internal/models"
	"github.com/stretchr/testify/assert"
)

func TestClient_DecodeReaderIdentityCallback(t *testing.T) {
	expectedSig := "249d9a838e9b981935324b02367ca72552aa430fc766f45f77fab7a81f9f3b9d"
	logger.Init(configs.Log{})
	log := logger.GetLogger()
	client := New(&configs.Idenfy{
		CallbackSignKey: "TestingKey",
	}, log)

	assert.NotNil(t, client, "Client is nil")
	webhook1, err := os.ReadFile("testdata/webhook.1.json")
	assert.NoError(t, err, "Could not open test data")
	err = client.VerifyCallbackSignature(context.Background(), webhook1, expectedSig)
	assert.NoError(t, err)
	var resp models.Verification
	decoder := json.NewDecoder(bytes.NewReader(webhook1))
	err = decoder.Decode(&resp)
	assert.NoError(t, err)
	// Basic verification info
	log.Info("resp", logger.Fields{
		"resp": resp,
	})
	assert.Equal(t, "123", resp.ClientID)
	assert.Equal(t, "scan-ref", resp.IdenfyRef)
	assert.Equal(t, "external-ref", resp.ExternalRef)
	assert.Equal(t, models.Platform("MOBILE_APP"), resp.Platform)
	assert.Equal(t, int64(1554726960), resp.StartTime)
	assert.Equal(t, int64(1554727002), resp.FinishTime)
	assert.Equal(t, "192.0.2.0", resp.ClientIP)
	assert.Equal(t, "LT", resp.ClientIPCountry)
	assert.Equal(t, "Kaunas, Lithuania", resp.ClientLocation)
	assert.False(t, *resp.Final)

	// Status checks
	assert.Equal(t, models.Overall("APPROVED"), *resp.Status.Overall)
	assert.Empty(t, resp.Status.SuspicionReasons)
	assert.Empty(t, resp.Status.MismatchTags)
	assert.Empty(t, resp.Status.FraudTags)
	assert.Equal(t, "DOC_VALIDATED", resp.Status.AutoDocument)
	assert.Equal(t, "FACE_MATCH", resp.Status.AutoFace)
	assert.Equal(t, "DOC_VALIDATED", resp.Status.ManualDocument)
	assert.Equal(t, "FACE_MATCH", resp.Status.ManualFace)
	assert.Nil(t, resp.Status.AdditionalSteps)

	// Document data
	assert.Equal(t, "FIRST-NAME-EXAMPLE", resp.Data.DocFirstName)
	assert.Equal(t, "LAST-NAME-EXAMPLE", resp.Data.DocLastName)
	assert.Equal(t, "XXXXXXXXX", resp.Data.DocNumber)
	assert.Equal(t, "XXXXXXXXX", resp.Data.DocPersonalCode)
	assert.Equal(t, "YYYY-MM-DD", resp.Data.DocExpiry)
	assert.Equal(t, "YYYY-MM-DD", resp.Data.DocDOB)
	assert.Equal(t, "2018-03-02", resp.Data.DocDateOfIssue)
	assert.Equal(t, models.DocumentType("ID_CARD"), *resp.Data.DocType)
	assert.Equal(t, models.Sex("UNDEFINED"), *resp.Data.DocSex)
	assert.Equal(t, "LT", resp.Data.DocNationality)
	assert.Equal(t, "LT", resp.Data.DocIssuingCountry)
	assert.Equal(t, "BIRTH PLACE", resp.Data.BirthPlace)
	assert.Equal(t, "AUTHORITY EXAMPLE", resp.Data.Authority)
	assert.Equal(t, "ADDRESS EXAMPLE", resp.Data.Address)
	assert.Equal(t, "FULL-NAME-EXAMPLE", resp.Data.FullName)
	assert.Equal(t, "LT", resp.Data.SelectedCountry)
	assert.False(t, *resp.Data.ManuallyDataChanged)
	assert.Equal(t, models.AgeEstimate("OVER_25"), *resp.Data.AgeEstimate)
	assert.Equal(t, "LOW", resp.Data.ClientIPProxyRiskLevel)

	// Original data
	assert.Equal(t, "FIRST-NAME-EXAMPLE", resp.Data.OrgFirstName)
	assert.Equal(t, "LAST-NAME-EXAMPLE", resp.Data.OrgLastName)
	assert.Equal(t, "LIETUVOS", resp.Data.OrgNationality)
	assert.Equal(t, "ŠILUVA", resp.Data.OrgBirthPlace)

	// File URLs
	expectedURLs := map[string]string{
		"FRONT": "https://s3.eu-west-1.amazonaws.com/production.users.storage/users_storage/users/<HASH>/FRONT.png?AWSAccessKeyId=<KEY>&Signature=<SIG>&Expires=<STAMP>",
		"BACK":  "https://s3.eu-west-1.amazonaws.com/production.users.storage/users_storage/users/<HASH>/BACK.png?AWSAccessKeyId=<KEY>&Signature=<SIG>&Expires=<STAMP>",
		"FACE":  "https://s3.eu-west-1.amazonaws.com/production.users.storage/users_storage/users/<HASH>/FACE.png?AWSAccessKeyId=<KEY>&Signature=<SIG>&Expires=<STAMP>",
	}
	assert.Equal(t, expectedURLs, resp.FileUrls)

	// AML and LID checks
	assert.Len(t, resp.AML, 1)
	assert.Equal(t, "PilotApiAmlV2", resp.AML[0].ServiceName)
	assert.Equal(t, "AML", resp.AML[0].ServiceGroupType)
	assert.Equal(t, "OHT8GR5ESRF5XROWE5ZGCC123", resp.AML[0].UID)
	assert.True(t, *resp.AML[0].Status.CheckSuccessful)
	assert.Equal(t, "NOT_SUSPECTED", resp.AML[0].Status.OverallStatus)

	assert.Len(t, resp.LID, 1)
	assert.Equal(t, "IrdInvalidPapers", resp.LID[0].ServiceName)
	assert.Equal(t, "LID", resp.LID[0].ServiceGroupType)
	assert.Equal(t, "OHT8GR5ESRF5XROWE5ZGCC123", resp.LID[0].UID)
	assert.True(t, *resp.LID[0].Status.CheckSuccessful)
	assert.Equal(t, "NOT_SUSPECTED", resp.LID[0].Status.OverallStatus)

	// Additional data
	assert.Empty(t, resp.AdditionalStepPdfUrls)
}
