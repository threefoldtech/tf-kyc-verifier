package middleware

import (
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/threefoldtech/tf-kyc-verifier/internal/config"
	"github.com/threefoldtech/tf-kyc-verifier/internal/errors"
	"github.com/threefoldtech/tf-kyc-verifier/internal/handlers"
	"github.com/threefoldtech/tf-kyc-verifier/internal/logger"
	"github.com/vedhavyas/go-subkey/v2"
	"github.com/vedhavyas/go-subkey/v2/ed25519"
	"github.com/vedhavyas/go-subkey/v2/sr25519"
)

// CORS returns a CORS middleware
func CORS() fiber.Handler {
	return cors.New()
}

// AuthMiddleware is a middleware that validates the authentication credentials
func AuthMiddleware(config config.Challenge) fiber.Handler {
	return func(c *fiber.Ctx) error {
		clientID := c.Get("X-Client-ID")
		signature := c.Get("X-Signature")
		challenge := c.Get("X-Challenge")

		if clientID == "" || signature == "" || challenge == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Missing authentication credentials",
			})
		}

		// Verify the clientID and signature here
		err := ValidateChallenge(clientID, signature, challenge, config.Domain, config.Window)
		if err != nil {
			// cast error to service error and convert it to http status code
			serviceError, ok := err.(*errors.ServiceError)
			if ok {
				return handlers.HandleServiceError(c, serviceError)
			}
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": err.Error(),
			})
		}
		// Verify the signature
		err = VerifySubstrateSignature(clientID, signature, challenge)
		if err != nil {
			serviceError, ok := err.(*errors.ServiceError)
			if ok {
				return handlers.HandleServiceError(c, serviceError)
			}
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": err.Error(),
			})
		}

		return c.Next()
	}
}

func fromHex(hex string) ([]byte, bool) {
	return subkey.DecodeHex(hex)
}

func VerifySubstrateSignature(address, signature, challenge string) error {
	challengeBytes, success := fromHex(challenge)
	if !success {
		return errors.NewValidationError("malformed challenge: failed to decode hex-encoded challenge", nil)
	}
	// hex to string
	sig, success := fromHex(signature)
	if !success {
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
		return errors.NewValidationError("error: can't create ed25519 public key", err)
	}

	if !pubkeyEd25519.Verify(challengeBytes, sig) {
		// Create a new sr25519 public key
		pubkeySr25519, err := sr25519.Scheme{}.FromPublicKey(pubkeyBytes)
		if err != nil {
			return errors.NewValidationError("error: can't create sr25519 public key", err)
		}
		if !pubkeySr25519.Verify(challengeBytes, sig) {
			return errors.NewAuthorizationError("bad signature: signature does not match", nil)
		}
	}

	return nil
}

func ValidateChallenge(address, signature, challenge, expectedDomain string, challengeWindow int64) error {
	// Parse and validate the challenge
	challengeBytes, success := fromHex(challenge)
	if !success {
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

func NewLoggingMiddleware(log logger.Logger) fiber.Handler {
	return func(c *fiber.Ctx) error {
		start := time.Now()
		path := c.Path()
		method := c.Method()
		ip := c.IP()

		// Log request
		log.Info("Incoming request", logger.Fields{
			"method":     method,
			"path":       path,
			"queries":    c.Queries(),
			"ip":         ip,
			"user_agent": string(c.Request().Header.UserAgent()),
			"headers":    c.GetReqHeaders(),
		})

		// Handle request
		err := c.Next()

		// Calculate duration
		duration := time.Since(start)
		status := c.Response().StatusCode()

		// Get response size
		responseSize := len(c.Response().Body())

		// Log the response
		logFields := logger.Fields{
			"method":        method,
			"path":          path,
			"ip":            ip,
			"status":        status,
			"duration":      duration,
			"response_size": responseSize,
		}

		// Add error if present
		if err != nil {
			logFields["error"] = err
			if status >= 500 {
				log.Error("Request failed", logFields)
			} else {
				log.Info("Request failed", logFields)
			}
		} else {
			log.Info("Request completed", logFields)
		}

		return err
	}
}
