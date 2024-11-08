package middleware

import (
	"fmt"
	"log/slog"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/threefoldtech/tf-kyc-verifier/internal/config"
	"github.com/threefoldtech/tf-kyc-verifier/internal/errors"
	"github.com/threefoldtech/tf-kyc-verifier/internal/handlers"
	"github.com/threefoldtech/tf-kyc-verifier/internal/responses"
	"github.com/vedhavyas/go-subkey/v2"
	"github.com/vedhavyas/go-subkey/v2/ed25519"
	"github.com/vedhavyas/go-subkey/v2/sr25519"
)

// AuthMiddleware is a middleware that validates the authentication credentials
func AuthMiddleware(config config.Challenge) fiber.Handler {
	return func(c *fiber.Ctx) error {
		clientID := c.Get("X-Client-ID")
		signature := c.Get("X-Signature")
		challenge := c.Get("X-Challenge")

		if clientID == "" || signature == "" || challenge == "" {
			return responses.RespondWithError(c, fiber.StatusBadRequest, fmt.Errorf("missing authentication credentials"))
		}

		// Verify the clientID and signature here
		err := ValidateChallenge(clientID, signature, challenge, config.Domain, config.Window)
		if err != nil {
			// cast error to service error and convert it to http status code
			serviceError, ok := err.(*errors.ServiceError)
			if ok {
				return handlers.HandleServiceError(c, serviceError)
			}
			return responses.RespondWithError(c, fiber.StatusBadRequest, err)
		}
		// Verify the signature
		err = VerifySubstrateSignature(clientID, signature, challenge)
		if err != nil {
			serviceError, ok := err.(*errors.ServiceError)
			if ok {
				return handlers.HandleServiceError(c, serviceError)
			}
			return responses.RespondWithError(c, fiber.StatusUnauthorized, err)
		}

		return c.Next()
	}
}

func fromHex(hex string) ([]byte, bool) {
	return subkey.DecodeHex(hex)
}

func VerifySubstrateSignature(address, signature, challenge string) error {
	challengeBytes, ok := fromHex(challenge)
	if !ok {
		return errors.NewValidationError("malformed challenge: failed to decode hex-encoded challenge", nil)
	}
	// hex to string
	sig, ok := fromHex(signature)
	if !ok {
		return errors.NewValidationError("malformed signature: failed to decode hex-encoded signature", nil)
	}
	// Convert address to public key
	_, pubkeyBytes, err := subkey.SS58Decode(address)
	if err != nil {
		return errors.NewValidationError("malformed address:failed to decode ss58 address", err)
	}

	// Create a new ed25519 public key
	pubkeyEd25519, err := ed25519.Scheme{}.FromPublicKey(pubkeyBytes)
	if err != nil {
		return errors.NewValidationError("creating ed25519 public key", err)
	}

	if !pubkeyEd25519.Verify(challengeBytes, sig) {
		// Create a new sr25519 public key
		pubkeySr25519, err := sr25519.Scheme{}.FromPublicKey(pubkeyBytes)
		if err != nil {
			return errors.NewValidationError("creating sr25519 public key", err)
		}
		if !pubkeySr25519.Verify(challengeBytes, sig) {
			return errors.NewAuthorizationError("bad signature: signature does not match", nil)
		}
	}

	return nil
}

func ValidateChallenge(address, signature, challenge, expectedDomain string, challengeWindow int64) error {
	// Parse and validate the challenge
	challengeBytes, ok := fromHex(challenge)
	if !ok {
		return errors.NewValidationError("malformed challenge: failed to decode hex-encoded challenge", nil)
	}
	parts := strings.Split(string(challengeBytes), ":")
	if len(parts) != 2 {
		return errors.NewValidationError("malformed challenge: invalid challenge format", nil)
	}

	// Check the domain
	if parts[0] != expectedDomain {
		return errors.NewValidationError("bad challenge: unexpected domain", nil)
	}

	// Check the timestamp
	timestamp, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil {
		return errors.NewValidationError("bad challenge: invalid timestamp", nil)
	}

	// Check if the timestamp is within an acceptable range (e.g., last 1 minutes)
	if time.Now().Unix()-timestamp > challengeWindow {
		return errors.NewValidationError("bad challenge: challenge expired", nil)
	}
	return nil
}

func NewLoggingMiddleware(logger *slog.Logger) fiber.Handler {
	return func(c *fiber.Ctx) error {
		start := time.Now()
		path := c.Path()
		method := c.Method()
		ip := c.IP()

		// Log request
		logger.Info("Incoming request", slog.Any("method", method), slog.Any("path", path), slog.Any("queries", c.Queries()), slog.Any("ip", ip), slog.Any("user_agent", string(c.Request().Header.UserAgent())), slog.Any("headers", c.GetReqHeaders()))

		// Handle request
		err := c.Next()

		// Calculate duration
		duration := time.Since(start)
		status := c.Response().StatusCode()

		// Get response size
		responseSize := len(c.Response().Body())

		// Log the response
		logger := logger.With(slog.Any("method", method), slog.Any("path", path), slog.Any("ip", ip), slog.Any("status", status), slog.Any("duration", duration), slog.Any("response_size", responseSize))

		// Add error if present
		if err != nil {
			logger = logger.With(slog.Any("error", err))
			if status >= 500 {
				logger.Error("Request failed")
			} else {
				logger.Info("Request failed")
			}
		} else {
			logger.Info("Request completed")
		}

		return err
	}
}
