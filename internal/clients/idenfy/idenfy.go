package idenfy

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"

	"example.com/tfgrid-kyc-service/internal/logger"
	"example.com/tfgrid-kyc-service/internal/models"
	"github.com/valyala/fasthttp"
)

type Idenfy struct {
	client *fasthttp.Client // TODO: Interface
	config IdenfyConfig     // TODO: Interface
	logger logger.Logger
}

const (
	VerificationSessionEndpoint = "/api/v2/token"
)

func New(config IdenfyConfig, logger logger.Logger) *Idenfy {
	return &Idenfy{
		client: &fasthttp.Client{},
		config: config,
		logger: logger,
	}
}

func (c *Idenfy) CreateVerificationSession(ctx context.Context, clientID string) (models.Token, error) { // TODO: Refactor
	url := c.config.GetBaseURL() + VerificationSessionEndpoint

	req := fasthttp.AcquireRequest()
	defer fasthttp.ReleaseRequest(req)

	req.SetRequestURI(url)
	req.Header.SetMethod(fasthttp.MethodPost)
	req.Header.Set("Content-Type", "application/json")

	// Set basic auth
	authStr := c.config.GetAPIKey() + ":" + c.config.GetAPISecret()
	auth := base64.StdEncoding.EncodeToString([]byte(authStr))
	req.Header.Set("Authorization", "Basic "+auth)

	RequestBody := c.createVerificationSessionRequestBody(clientID, c.config.GetDevMode())

	jsonBody, err := json.Marshal(RequestBody)
	if err != nil {
		return models.Token{}, fmt.Errorf("error marshaling request body: %w", err)
	}
	req.SetBody(jsonBody)

	resp := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseResponse(resp)
	c.logger.Debug("Preparing iDenfy verification session request", map[string]interface{}{
		"request": jsonBody,
	})
	err = c.client.Do(req, resp)
	if err != nil {
		return models.Token{}, fmt.Errorf("error sending request: %w", err)
	}

	if resp.StatusCode() < 200 || resp.StatusCode() >= 300 {
		c.logger.Debug("Received unexpected status code from iDenfy", map[string]interface{}{
			"status": resp.StatusCode(),
			"error":  string(resp.Body()),
		})
		return models.Token{}, fmt.Errorf("unexpected status code: %d", resp.StatusCode())
	}
	c.logger.Debug("Received response from iDenfy", map[string]interface{}{
		"response": string(resp.Body()),
	})

	var result models.Token
	if err := json.Unmarshal(resp.Body(), &result); err != nil {
		return models.Token{}, fmt.Errorf("error decoding response: %w", err)
	}

	return result, nil
}

// verify signature of the callback
func (c *Idenfy) VerifyCallbackSignature(ctx context.Context, body []byte, sigHeader string) error {
	if len(c.config.GetCallbackSignKey()) < 1 {
		return errors.New("callback was received but no signature key was provided")
	}
	sig, err := hex.DecodeString(sigHeader)
	if err != nil {
		return err
	}
	mac := hmac.New(sha256.New, []byte(c.config.GetCallbackSignKey()))

	mac.Write(body)

	if !hmac.Equal(sig, mac.Sum(nil)) {
		return errors.New("signature verification failed")
	}
	return nil
}

// function to create a request body for the verification session
func (c *Idenfy) createVerificationSessionRequestBody(clientID string, devMode bool) map[string]interface{} {
	RequestBody := map[string]interface{}{
		"clientId":            clientID,
		"generateDigitString": true,
		"callbackUrl":         c.config.GetCallbackUrl(),
	}
	if devMode {
		RequestBody["expiryTime"] = 30
		RequestBody["dummyStatus"] = "APPROVED"
	}
	return RequestBody
}
