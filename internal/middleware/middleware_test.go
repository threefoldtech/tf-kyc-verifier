package middleware

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	io_prometheus_client "github.com/prometheus/client_model/go"
	"github.com/stretchr/testify/assert"
	"github.com/threefoldtech/tf-kyc-verifier/internal/config"
	"github.com/threefoldtech/tf-kyc-verifier/internal/metrics"
	"github.com/vedhavyas/go-subkey/v2"
	"github.com/vedhavyas/go-subkey/v2/ed25519"
	"github.com/vedhavyas/go-subkey/v2/sr25519"
)

func TestAuthMiddleware(t *testing.T) {
	// Setup
	app := fiber.New()
	cfg := config.Challenge{
		Window: 8,
		Domain: "test.grid.tf",
	}

	// Mock handler that should be called after middleware
	successHandler := func(c *fiber.Ctx) error {
		return c.SendStatus(fiber.StatusOK)
	}

	// Apply middleware
	app.Use(AuthMiddleware(cfg))
	app.Get("/test", successHandler)

	// Generate keys
	krSr25519, err := generateTestSr25519Keys()
	if err != nil {
		t.Fatal(err)
	}
	krEd25519, err := generateTestEd25519Keys()
	if err != nil {
		t.Fatal(err)
	}
	clientIDSr := krSr25519.SS58Address(42)
	clientIDEd := krEd25519.SS58Address(42)
	invalidChallenge := createInvalidSignMessageInvalidFormat(cfg.Domain)
	expiredChallenge := createInvalidSignMessageExpired(cfg.Domain)
	wrongDomainChallenge := createInvalidSignMessageWrongDomain()
	validChallenge := createValidSignMessage(cfg.Domain)
	sigSr, err := krSr25519.Sign([]byte(validChallenge))
	if err != nil {
		t.Fatal(err)
	}
	sigEd, err := krEd25519.Sign([]byte(validChallenge))
	if err != nil {
		t.Fatal(err)
	}
	sigSrHex := hex.EncodeToString(sigSr)
	sigEdHex := hex.EncodeToString(sigEd)
	tests := []struct {
		name           string
		clientID       string
		signature      string
		challenge      string
		expectedStatus int
		expectedError  string
	}{
		{
			name:           "Missing all credentials",
			clientID:       "",
			signature:      "",
			challenge:      "",
			expectedStatus: fiber.StatusBadRequest,
			expectedError:  "missing authentication credentials",
		},
		{
			name:           "Missing client ID",
			clientID:       "",
			signature:      sigSrHex,
			challenge:      toHex(validChallenge),
			expectedStatus: fiber.StatusBadRequest,
			expectedError:  "missing authentication credentials",
		},
		{
			name:           "Missing signature",
			clientID:       clientIDSr,
			signature:      "",
			challenge:      toHex(validChallenge),
			expectedStatus: fiber.StatusBadRequest,
			expectedError:  "missing authentication credentials",
		},
		{
			name:           "Missing challenge",
			clientID:       clientIDSr,
			signature:      sigSrHex,
			challenge:      "",
			expectedStatus: fiber.StatusBadRequest,
			expectedError:  "missing authentication credentials",
		},
		{
			name:           "Invalid client ID format",
			clientID:       toHex("invalid_client_id"),
			signature:      sigSrHex,
			challenge:      toHex(validChallenge),
			expectedStatus: fiber.StatusBadRequest,
			expectedError:  "malformed address",
		},
		{
			name:           "Invalid challenge format",
			clientID:       clientIDSr,
			signature:      sigSrHex,
			challenge:      toHex(invalidChallenge),
			expectedStatus: fiber.StatusBadRequest,
			expectedError:  "invalid challenge format",
		},
		{
			name:           "Expired challenge",
			clientID:       clientIDSr,
			signature:      sigSrHex,
			challenge:      toHex(expiredChallenge),
			expectedStatus: fiber.StatusBadRequest,
			expectedError:  "challenge expired",
		},
		{
			name:           "Invalid domain in challenge",
			clientID:       clientIDSr,
			signature:      sigSrHex,
			challenge:      toHex(wrongDomainChallenge),
			expectedStatus: fiber.StatusBadRequest,
			expectedError:  "unexpected domain",
		},
		{
			name:           "invalid signature format",
			clientID:       clientIDSr,
			signature:      "invalid_signature",
			challenge:      toHex(validChallenge),
			expectedStatus: fiber.StatusBadRequest,
			expectedError:  "malformed signature",
		},
		{
			name:           "bad signature",
			clientID:       clientIDSr,
			signature:      sigEdHex,
			challenge:      toHex(validChallenge),
			expectedStatus: fiber.StatusUnauthorized,
			expectedError:  "signature does not match",
		},
		{
			name:           "valid credentials SR25519",
			clientID:       clientIDSr,
			signature:      sigSrHex,
			challenge:      toHex(validChallenge),
			expectedStatus: fiber.StatusOK,
		},
		{
			name:           "valid credentials ED25519",
			clientID:       clientIDEd,
			signature:      sigEdHex,
			challenge:      toHex(validChallenge),
			expectedStatus: fiber.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create request
			req := createTestRequest(tt.clientID, tt.signature, tt.challenge)
			resp, err := app.Test(req)

			// Assert response
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedStatus, resp.StatusCode)

			// Check error message if expected
			if tt.expectedError != "" {
				var errorResp struct {
					Error string `json:"error"`
				}
				err = parseResponse(resp, &errorResp)
				assert.NoError(t, err)
				assert.Contains(t, errorResp.Error, tt.expectedError)
			}
		})
	}
}

// Helper function to create test requests
func createTestRequest(clientID, signature, challenge string) *http.Request {
	req := httptest.NewRequest(fiber.MethodGet, "/test", nil)
	if clientID != "" {
		req.Header.Set("X-Client-ID", clientID)
	}
	if signature != "" {
		req.Header.Set("X-Signature", signature)
	}
	if challenge != "" {
		req.Header.Set("X-Challenge", challenge)
	}
	return req
}

// Helper function to parse response body
func parseResponse(resp *http.Response, v interface{}) error {
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	return json.Unmarshal(body, v)
}

func toHex(message string) string {
	return hex.EncodeToString([]byte(message))
}

func createValidSignMessage(domain string) string {
	// return a message with the domain and the current timestamp in hex
	message := fmt.Sprintf("%s:%d", domain, time.Now().Unix())
	return message
}

func createInvalidSignMessageWrongDomain() string {
	// return a message with the domain and the current timestamp in hex
	message := fmt.Sprintf("%s:%d", "wrong.domain", time.Now().Unix())
	return message
}

func createInvalidSignMessageExpired(domain string) string {
	// return a message with the domain and the current timestamp in hex
	message := fmt.Sprintf("%s:%d", domain, time.Now().Add(-10*time.Minute).Unix())
	return message
}

func createInvalidSignMessageInvalidFormat(domain string) string {
	// return a message with the domain and the current timestamp in hex
	message := fmt.Sprintf("%s%d", domain, time.Now().Unix())
	return message
}

func generateTestSr25519Keys() (subkey.KeyPair, error) {
	krSr25519, err := sr25519.Scheme{}.Generate()
	if err != nil {
		return nil, err
	}
	return krSr25519, nil
}

func generateTestEd25519Keys() (subkey.KeyPair, error) {
	krEd25519, err := ed25519.Scheme{}.Generate()
	if err != nil {
		return nil, err
	}
	return krEd25519, nil
}

func TestMetricsMiddleware(t *testing.T) {
	// Setup
	app := fiber.New()

	// Mock handler that should be called after middleware
	successHandler := func(c *fiber.Ctx) error {
		return c.SendStatus(fiber.StatusOK)
	}

	failHandler := func(c *fiber.Ctx) error {
		return c.SendStatus(fiber.StatusInternalServerError)
	}

	m := metrics.GetInstance()
	err := m.Register()
	assert.NoError(t, err)

	app.Use(MetricsMiddleware(m))
	app.Get("/test_success", successHandler)
	app.Get("/test_fail", failHandler)

	t.Run("successful count for recieved requests", func(t *testing.T) {

		req := httptest.NewRequest(fiber.MethodGet, "/test_success", nil)
		_, err := app.Test(req)

		assert.NoError(t, err)
		metric, err := m.HTTPRequestsReceived.GetMetricWithLabelValues("GET", "/test_success")

		assert.NoError(t, err)

		got := &io_prometheus_client.Metric{}
		err = metric.Write(got)

		assert.NoError(t, err)
		assert.Equal(t, float64(1), got.GetCounter().GetValue())

	})

	t.Run("successful count for failed requests", func(t *testing.T) {

		req := httptest.NewRequest(fiber.MethodGet, "/test_fail", nil)
		_, err := app.Test(req)

		assert.NoError(t, err)

		metric1, err1 := m.HTTPRequestsReceived.GetMetricWithLabelValues("GET", "/test_fail")
		metric2, err2 := m.InternalServerErrorRate.GetMetricWithLabelValues("GET", "/test_fail")

		assert.NoError(t, err1)
		assert.NoError(t, err2)

		got1 := &io_prometheus_client.Metric{}
		got2 := &io_prometheus_client.Metric{}

		err = metric1.Write(got1)
		assert.NoError(t, err)
		assert.Equal(t, float64(1), got1.GetCounter().GetValue())

		err = metric2.Write(got2)
		assert.NoError(t, err)
		assert.Equal(t, float64(1), got2.GetGauge().GetValue())

	})
}
