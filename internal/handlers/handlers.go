package handlers

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"time"

	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"

	"example.com/tfgrid-kyc-service/internal/errors"
	"example.com/tfgrid-kyc-service/internal/logger"
	"example.com/tfgrid-kyc-service/internal/models"
	"example.com/tfgrid-kyc-service/internal/responses"
	"example.com/tfgrid-kyc-service/internal/services"
)

type Handler struct {
	kycService services.KYCService
	logger     *logger.Logger
}

func NewHandler(kycService services.KYCService, logger *logger.Logger) *Handler {
	return &Handler{kycService: kycService, logger: logger}
}

// @Summary		Get or Generate iDenfy Verification Token
// @Description	Returns a token for a client
// @Tags			Token
// @Accept			json
// @Produce		json
// @Param			X-Client-ID	header		string	true	"TFChain SS58Address"								minlength(48)	maxlength(48)
// @Param			X-Challenge	header		string	true	"hex-encoded message `{api-domain}:{timestamp}`"
// @Param			X-Signature	header		string	true	"hex-encoded sr25519|ed25519 signature"				minlength(128)	maxlength(128)
// @Success		200			{object}	responses.TokenResponse "Existing token retrieved"
// @Success		201			{object}	responses.TokenResponse "New token created"
// @Failure		401			{object}	responses.ErrorResponse
// @Failure		402			{object}	responses.ErrorResponse
// @Failure		409			{object}	responses.ErrorResponse
// @Failure		500			{object}	responses.ErrorResponse
// @Router			/api/v1/token [post]
func (h *Handler) GetorCreateVerificationToken() fiber.Handler {
	return func(c *fiber.Ctx) error {
		clientID := c.Get("X-Client-ID")

		token, isNewToken, err := h.kycService.GetorCreateVerificationToken(c.Context(), clientID)
		if err != nil {
			return handleError(c, err)
		}
		response := responses.NewTokenResponseWithStatus(token, isNewToken)
		if isNewToken {
			return c.Status(fiber.StatusCreated).JSON(fiber.Map{"result": response})
		}
		return c.Status(fiber.StatusOK).JSON(fiber.Map{"result": response})
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
// @Success		200			{object}		responses.VerificationDataResponse
// @Failure		401			{object}		responses.ErrorResponse
// @Failure		404			{object}		responses.ErrorResponse
// @Failure		500			{object}		responses.ErrorResponse
// @Router			/api/v1/data [get]
func (h *Handler) GetVerificationData() fiber.Handler {
	return func(c *fiber.Ctx) error {
		clientID := c.Get("X-Client-ID")
		verification, err := h.kycService.GetVerificationData(c.Context(), clientID)
		if err != nil {
			return handleError(c, err)
		}
		if verification == nil {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Verification not found"})
		}
		response := responses.NewVerificationDataResponse(verification)
		return c.JSON(fiber.Map{"result": response})
	}
}

// @Summary		Get Verification Status
// @Description	Returns the verification status for a client
// @Tags			Verification
// @Accept			json
// @Produce		json
// @Param			client_id	query		string	false	"TFChain SS58Address"								minlength(48)	maxlength(48)
// @Param			twin_id		query		string	false	"Twin ID"											minlength(1)
// @Success		200			{object}		responses.VerificationStatusResponse
// @Failure		400			{object}		responses.ErrorResponse
// @Failure		404			{object}		responses.ErrorResponse
// @Failure		500			{object}		responses.ErrorResponse
// @Router			/api/v1/status [get]
func (h *Handler) GetVerificationStatus() fiber.Handler {
	return func(c *fiber.Ctx) error {
		clientID := c.Query("client_id")
		twinID := c.Query("twin_id")

		if clientID == "" && twinID == "" {
			h.logger.Warn("Bad request: missing client_id and twin_id")
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Either client_id or twin_id must be provided"})
		}
		var verification *models.VerificationOutcome
		var err error

		if clientID != "" {
			verification, err = h.kycService.GetVerificationStatus(c.Context(), clientID)
		} else {
			verification, err = h.kycService.GetVerificationStatusByTwinID(c.Context(), twinID)
		}
		if err != nil {
			h.logger.Error("Failed to get verification status",
				zap.String("clientID", clientID),
				zap.String("twinID", twinID),
				zap.Error(err),
			)
			return handleError(c, err)
		}
		if verification == nil {
			h.logger.Info("Verification not found",
				zap.String("clientID", clientID),
				zap.String("twinID", twinID),
			)
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Verification not found"})
		}
		response := responses.NewVerificationStatusResponse(verification)
		return c.JSON(fiber.Map{"result": response})
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
		// decode base64 to string
		dst := make([]byte, base64.StdEncoding.DecodedLen(len(c.Body())))
		base64.StdEncoding.Decode(dst, c.Body())
		h.logger.Debug("Received verification update", zap.Any("body", string(dst)), zap.Any("headers", &c.Request().Header))

		sigHeader := c.Get("Idenfy-Signature")
		if len(sigHeader) < 1 {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "No signature provided"})
		}
		body := c.Body()
		var result models.Verification
		decoder := json.NewDecoder(bytes.NewReader(body))
		err := decoder.Decode(&result)
		if err != nil {
			h.logger.Error("Error decoding verification update", zap.Error(err))
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
		}
		h.logger.Debug("Verification update after decoding", zap.Any("result", result))
		err = h.kycService.ProcessVerificationResult(c.Context(), body, sigHeader, result)
		if err != nil {
			return handleError(c, err)
		}
		return c.SendStatus(fiber.StatusOK)
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
		// TODO: implement
		h.logger.Error("Received ID expiration notification but not implemented")
		return c.SendStatus(fiber.StatusNotImplemented)
	}
}

// @Summary		Health Check
// @Description	Returns the health status of the service
// @Tags			Health
// @Success		200	{object}	responses.HealthResponse
// @Router			/api/v1/health [get]
func (h *Handler) HealthCheck() fiber.Handler {
	return func(c *fiber.Ctx) error {
		health := responses.HealthResponse{
			Status:    "ok",
			Timestamp: time.Now().UTC().Format(time.RFC3339),
		}

		return c.JSON(health)
	}
}

func handleError(c *fiber.Ctx, err error) error {
	if serviceErr, ok := err.(*errors.ServiceError); ok {
		return handleServiceError(c, serviceErr)
	}
	return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
}

func handleServiceError(c *fiber.Ctx, err *errors.ServiceError) error {
	statusCode := getStatusCode(err.Type)
	return c.Status(statusCode).JSON(fiber.Map{
		"error": err.Message,
	})
}

func getStatusCode(errorType errors.ErrorType) int {
	switch errorType {
	case errors.ErrorTypeValidation:
		return fiber.StatusBadRequest
	case errors.ErrorTypeAuthorization:
		return fiber.StatusUnauthorized
	case errors.ErrorTypeNotFound:
		return fiber.StatusNotFound
	case errors.ErrorTypeConflict:
		return fiber.StatusConflict
	case errors.ErrorTypeExternal:
		return fiber.StatusInternalServerError
	case errors.ErrorTypeNotSufficientBalance:
		return fiber.StatusPaymentRequired
	default:
		return fiber.StatusInternalServerError
	}
}
