/*
Package idenfy contains the iDenfy client for the application.
This layer is responsible for interacting with the iDenfy API. the main operations are:
- creating a verification session
- verifying the callback signature
*/
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
	"log/slog"
	"time"

	"github.com/threefoldtech/tf-kyc-verifier/internal/models"
	"github.com/valyala/fasthttp"
)

type Idenfy struct {
	client *fasthttp.Client // TODO: Interface
	config IdenfyConfig     // TODO: Interface
	logger *slog.Logger
}

const (
	VerificationSessionEndpoint = "/api/v2/token"
)

func New(config IdenfyConfig, logger *slog.Logger) *Idenfy {
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
		return models.Token{}, fmt.Errorf("marshaling request body: %w", err)
	}
	req.SetBody(jsonBody)
	// Set deadline from context
	deadline, ok := ctx.Deadline()
	if ok {
		req.SetTimeout(time.Until(deadline))
	}

	resp := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseResponse(resp)
	c.logger.Debug("Preparing iDenfy verification session request", "request", jsonBody)
	err = c.client.Do(req, resp)
	if err != nil {
		return models.Token{}, fmt.Errorf("sending token request to iDenfy: %w", err)
	}

	if resp.StatusCode() < 200 || resp.StatusCode() >= 300 {
		c.logger.Debug("Received unexpected status code from iDenfy", "status", resp.StatusCode(), "error", string(resp.Body()))
		return models.Token{}, fmt.Errorf("unexpected status code from iDenfy: %d", resp.StatusCode())
	}
	c.logger.Debug("Received response from iDenfy", "response", string(resp.Body()))

	var result models.Token
	if err := json.Unmarshal(resp.Body(), &result); err != nil {
		return models.Token{}, fmt.Errorf("decoding token response from iDenfy: %w", err)
	}

	return result, nil
}

// verify signature of the callback
func (c *Idenfy) VerifyCallbackSignature(ctx context.Context, body []byte, sigHeader string) error {
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
