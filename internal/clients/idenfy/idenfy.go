/*
Package idenfy contains the iDenfy client for the application.
This layer is responsible for interacting with the iDenfy API. the main operations are:
- creating a verification session
- verifying the callback signature
*/
package idenfy

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"time"

	"github.com/threefoldtech/tf-kyc-verifier/internal/models"
)

type Idenfy struct {
	client *http.Client
	config IdenfyConfig
	logger *slog.Logger
}

const (
	VerificationSessionEndpoint = "/api/v2/token"
	TokenExpirySeconds          = 86400
	TokenExpiryDevModeSeconds   = 30
	DefaultTimeout              = 10 * time.Second
	ContentTypeJSON             = "application/json"
)

func New(config IdenfyConfig, logger *slog.Logger) *Idenfy {
	return &Idenfy{
		client: &http.Client{
			Timeout: DefaultTimeout,
		},
		config: config,
		logger: logger,
	}
}

func (c *Idenfy) CreateVerificationSession(ctx context.Context, clientID string) (models.Token, error) {
	req, err := c.prepareRequest(ctx, clientID, TokenExpirySeconds)
	if err != nil {
		return models.Token{}, fmt.Errorf("preparing request: %w", err)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return models.Token{}, fmt.Errorf("sending request: %w", err)
	}
	defer resp.Body.Close()

	return c.handleResponse(resp)
}

func (c *Idenfy) VerifyCallbackSignature(ctx context.Context, body []byte, sigHeader string) error {
	sig, err := hex.DecodeString(sigHeader)
	if err != nil {
		return fmt.Errorf("invalid signature format: %w", err)
	}
	mac := hmac.New(sha256.New, []byte(c.config.GetCallbackSignKey()))
	mac.Write(body)

	if !hmac.Equal(sig, mac.Sum(nil)) {
		return errors.New("signature verification failed")
	}
	return nil
}

func (c *Idenfy) prepareRequest(ctx context.Context, clientID string, tokenExpiryTime int) (*http.Request, error) {
	body, err := c.createRequestBody(clientID, tokenExpiryTime)
	if err != nil {
		return nil, fmt.Errorf("creating request body: %w", err)
	}

	url := c.config.GetBaseURL() + VerificationSessionEndpoint
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	c.setRequestHeaders(req)
	return req, nil
}

func (c *Idenfy) createRequestBody(clientID string, tokenExpiryTime int) ([]byte, error) {
	requestBody := c.createVerificationSessionRequestBody(clientID, c.config.GetDevMode(), tokenExpiryTime)
	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("marshaling request body: %w", err)
	}
	return jsonBody, nil
}

func (c *Idenfy) setRequestHeaders(req *http.Request) {
	req.Header.Set("Content-Type", ContentTypeJSON)
	authStr := c.config.GetAPIKey() + ":" + c.config.GetAPISecret()
	auth := base64.StdEncoding.EncodeToString([]byte(authStr))
	req.Header.Set("Authorization", "Basic "+auth)
}

func (c *Idenfy) handleResponse(resp *http.Response) (models.Token, error) {
	if err := c.validateResponseStatus(resp); err != nil {
		return models.Token{}, err
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return models.Token{}, fmt.Errorf("reading response body: %w", err)
	}

	var result models.Token
	if err := json.Unmarshal(body, &result); err != nil {
		return models.Token{}, fmt.Errorf("decoding token response from iDenfy: %w", err)
	}

	return result, nil
}

func (c *Idenfy) validateResponseStatus(resp *http.Response) error {
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		c.logger.Debug("Received unexpected status code from iDenfy",
			"status", resp.StatusCode,
			"error", string(body),
		)
		return fmt.Errorf("unexpected status code from iDenfy: code: %d, body: %s", resp.StatusCode, string(body))
	}
	return nil
}

type VerificationSessionRequest struct {
	ClientID            string `json:"clientId"`
	GenerateDigitString bool   `json:"generateDigitString"`
	CallbackURL         string `json:"callbackUrl"`
	ExpiryTime          int    `json:"expiryTime"`
	DummyStatus         string `json:"dummyStatus,omitempty"`
}

func (c *Idenfy) createVerificationSessionRequestBody(clientID string, devMode bool, tokenExpiryTime int) *VerificationSessionRequest {
	RequestBody := &VerificationSessionRequest{
		ClientID:            clientID,
		GenerateDigitString: true,
		CallbackURL:         c.config.GetCallbackUrl(),
		ExpiryTime:          tokenExpiryTime,
	}
	if devMode {
		RequestBody.ExpiryTime = TokenExpiryDevModeSeconds
		RequestBody.DummyStatus = "APPROVED"
	}
	c.logger.Debug("Creating verification session", "request", RequestBody)
	return RequestBody
}
