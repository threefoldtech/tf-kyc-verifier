package middleware

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/vedhavyas/go-subkey/v2"
	"github.com/vedhavyas/go-subkey/v2/ed25519"
	"github.com/vedhavyas/go-subkey/v2/sr25519"
)

// Logger returns a logger middleware
func Logger() fiber.Handler {
	return logger.New()
}

// CORS returns a CORS middleware
func CORS() fiber.Handler {
	return cors.New()
}

// AuthMiddleware is a middleware that validates the authentication credentials
func AuthMiddleware(challengeWindow int64) fiber.Handler {
	return func(c *fiber.Ctx) error {
		clientID := c.Get("X-Client-ID")
		signature := c.Get("X-Signature")
		challenge := c.Get("X-Challenge")

		if clientID == "" || signature == "" || challenge == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Missing authentication credentials",
			})
		}

		// Verify the clientID and signature here
		err := ValidateChallenge(clientID, signature, challenge, "kyc1.gent01.dev.grid.tf", challengeWindow)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": err.Error(),
			})
		}
		// Verify the signature
		err = VerifySubstrateSignature(clientID, signature, challenge)
		if err != nil {
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
		return fmt.Errorf("malformed challenge: failed to decode hex-encoded challenge")
	}
	// hex to string
	sig, success := fromHex(signature)
	if !success {
		return fmt.Errorf("malformed signature: failed to decode hex-encoded signature")
	}
	// Convert address to public key
	_, pubkeyBytes, err := subkey.SS58Decode(address)
	if err != nil {
		return fmt.Errorf("malformed address:failed to decode ss58 address: %w", err)
	}

	// Create a new ed25519 public key
	pubkeyEd25519, err := ed25519.Scheme{}.FromPublicKey(pubkeyBytes)
	if err != nil {
		return fmt.Errorf("error: can't create ed25519 public key: %w", err)
	}

	if !pubkeyEd25519.Verify(challengeBytes, sig) {
		// Create a new sr25519 public key
		pubkeySr25519, err := sr25519.Scheme{}.FromPublicKey(pubkeyBytes)
		if err != nil {
			return fmt.Errorf("error: can't create sr25519 public key: %w", err)
		}
		if !pubkeySr25519.Verify(challengeBytes, sig) {
			return fmt.Errorf("bad signature: signature does not match")
		}
	}

	return nil
}

func ValidateChallenge(address, signature, challenge, expectedDomain string, challengeWindow int64) error {
	// Parse and validate the challenge
	challengeBytes, success := fromHex(challenge)
	if !success {
		return fmt.Errorf("malformed challenge: failed to decode hex-encoded challenge")
	}
	parts := strings.Split(string(challengeBytes), ":")
	if len(parts) != 2 {
		return fmt.Errorf("malformed challenge: invalid challenge format")
	}

	// Check the domain
	if parts[0] != expectedDomain {
		return fmt.Errorf("bad challenge: unexpected domain")
	}

	// Check the timestamp
	timestamp, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil {
		return fmt.Errorf("bad challenge: invalid timestamp")
	}

	// Check if the timestamp is within an acceptable range (e.g., last 1 minutes)
	if time.Now().Unix()-timestamp > challengeWindow {
		return fmt.Errorf("bad challenge: challenge expired")
	}
	return nil
}
