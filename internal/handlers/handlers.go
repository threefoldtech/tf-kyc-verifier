package handlers

import (
	"fmt"

	"github.com/gofiber/fiber/v2"

	"example.com/tfgrid-kyc-service/internal/services"
)

type Handler struct {
	tokenService        services.TokenService
	verificationService services.VerificationService
}

func NewHandler(tokenService services.TokenService, verificationService services.VerificationService) *Handler {
	return &Handler{tokenService: tokenService, verificationService: verificationService}
}

// @Summary		Get or Generate iDenfy Verification Token
// @Description	Returns a token for a client
// @Tags			Token
// @Accept			json
// @Produce		json
// @Param			X-Client-ID	header		string	true	"TFChain SS58Address"								minlength(48)	maxlength(48)
// @Param			X-Challenge	header		string	true	"hex-encoded message `{api-domain}:{timestamp}`"
// @Param			X-Signature	header		string	true	"hex-encoded sr25519|ed25519 signature"				minlength(128)	maxlength(128)
// @Success		200			{object}	responses.TokenResponseWithStatus
// @Router			/api/v1/token [post]
func (h *Handler) GetorCreateVerificationToken() fiber.Handler {
	return func(c *fiber.Ctx) error {
		clientID := c.Get("X-Client-ID")
		// check if user account balance satisfies the minimum required balance, return an error if not
		hasRequiredBalance, err := h.tokenService.AccountHasRequiredBalance(c.Context(), clientID)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
		}
		if !hasRequiredBalance {
			return c.Status(fiber.StatusPaymentRequired).JSON(fiber.Map{"error": "Account does not have the required balance"})
		}
		// check if user is unverified, return an error if not
		// this should be client responsibility to check if they are verified before requesting a new verification
		isVerified, err := h.verificationService.IsUserVerified(c.Context(), clientID)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
		}
		if isVerified {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "User already verified"})
		}

		fmt.Println("creating new token")
		token, err := h.tokenService.GetorCreateVerificationToken(c.Context(), clientID)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
		}
		fmt.Println("token from handler", token)
		return c.JSON(fiber.Map{"result": token})
	}
}

// @Summary		Get Verification Data
// @Description	Returns the verification data for a client
// @Tags			Verification
// @Accept			json
// @Produce		json
// @Param			X-Client-ID	header		string	true	"TFChain SS58Address"								minlength(48)	maxlength(48)
// @Param			X-Challenge	header		string	true	"hex-encoded message `{api-domain}:{timestamp}`"
// @Param			X-Signature	header		string	true	"hex-encoded sr25519|ed25519 signature"				minlength(128)	maxlength(128)
// @Success		200			{object}	responses.VerificationDataResponse
// @Router			/api/v1/data [get]
func (h *Handler) GetVerificationData() fiber.Handler {
	return func(c *fiber.Ctx) error {
		clientID := c.Query("clientID")
		result, err := h.verificationService.GetVerificationData(c.Context(), clientID)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
		}
		if result == nil {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Verification not found"})
		}
		return c.JSON(fiber.Map{"result": result})
	}
}

// @Summary		Get Verification Status
// @Description	Returns the verification status for a client
// @Tags			Verification
// @Accept			json
// @Produce		json
// @Param			X-Client-ID	header		string	true	"TFChain SS58Address"								minlength(48)	maxlength(48)
// @Param			X-Challenge	header		string	true	"hex-encoded message `{api-domain}:{timestamp}`"
// @Param			X-Signature	header		string	true	"hex-encoded sr25519|ed25519 signature"				minlength(128)	maxlength(128)
// @Success		200			{object}	responses.VerificationStatusResponse
// @Router			/api/v1/status [get]
func (h *Handler) GetVerificationStatus() fiber.Handler {
	return func(c *fiber.Ctx) error {
		clientID := c.Query("clientID")
		result, err := h.verificationService.GetVerificationStatus(c.Context(), clientID)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
		}
		if result == nil {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Verification not found"})
		}
		return c.JSON(fiber.Map{"result": result})
	}
}

// @Summary		Process Verification Update
// @Description	Processes the verification update for a client
// @Tags			Webhooks
// @Accept			json
// @Produce		json
// @Success		200
// @Router			/webhooks/idenfy/verification-update [post]
func (h *Handler) ProcessVerificationResult() fiber.Handler {
	return func(c *fiber.Ctx) error {
		return nil
	}
}

// @Summary		Process Doc Expiration Notification
// @Description	Processes the doc expiration notification for a client
// @Tags			Webhooks
// @Accept			json
// @Produce		json
// @Success		200
// @Router			/webhooks/idenfy/id-expiration [post]
func (h *Handler) ProcessDocExpirationNotification() fiber.Handler {
	return func(c *fiber.Ctx) error {
		return nil
	}
}
