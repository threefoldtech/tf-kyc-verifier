package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"time"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/readpref"

	"example.com/tfgrid-kyc-service/internal/build"
	"example.com/tfgrid-kyc-service/internal/configs"
	"example.com/tfgrid-kyc-service/internal/errors"
	"example.com/tfgrid-kyc-service/internal/logger"
	"example.com/tfgrid-kyc-service/internal/models"
	"example.com/tfgrid-kyc-service/internal/responses"
	"example.com/tfgrid-kyc-service/internal/services"
)

type Handler struct {
	kycService services.KYCService
	config     *configs.Config
	logger     logger.Logger
}

func NewHandler(kycService services.KYCService, config *configs.Config, logger logger.Logger) *Handler {
	return &Handler{kycService: kycService, config: config, logger: logger}
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
// @Failure		400			{object}	responses.ErrorResponse
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
			return HandleError(c, err)
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
// @Failure		400			{object}		responses.ErrorResponse
// @Failure		401			{object}		responses.ErrorResponse
// @Failure		404			{object}		responses.ErrorResponse
// @Failure		500			{object}		responses.ErrorResponse
// @Router			/api/v1/data [get]
func (h *Handler) GetVerificationData() fiber.Handler {
	return func(c *fiber.Ctx) error {
		clientID := c.Get("X-Client-ID")
		verification, err := h.kycService.GetVerificationData(c.Context(), clientID)
		if err != nil {
			return HandleError(c, err)
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
			h.logger.Warn("Bad request: missing client_id and twin_id", map[string]interface{}{})
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
			h.logger.Error("Failed to get verification status", map[string]interface{}{
				"clientID": clientID,
				"twinID":   twinID,
				"error":    err,
			})
			return HandleError(c, err)
		}
		if verification == nil {
			h.logger.Info("Verification not found", map[string]interface{}{
				"clientID": clientID,
				"twinID":   twinID,
			})
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
		h.logger.Debug("Received verification update", map[string]interface{}{
			"body":    string(c.Body()),
			"headers": &c.Request().Header,
		})
		sigHeader := c.Get("Idenfy-Signature")
		if len(sigHeader) < 1 {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "No signature provided"})
		}
		body := c.Body()
		var result models.Verification
		decoder := json.NewDecoder(bytes.NewReader(body))
		err := decoder.Decode(&result)
		if err != nil {
			h.logger.Error("Error decoding verification update", map[string]interface{}{
				"error": err,
			})
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
		}
		h.logger.Debug("Verification update after decoding", map[string]interface{}{
			"result": result,
		})
		err = h.kycService.ProcessVerificationResult(c.Context(), body, sigHeader, result)
		if err != nil {
			return HandleError(c, err)
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
		h.logger.Error("Received ID expiration notification but not implemented", map[string]interface{}{})
		return c.SendStatus(fiber.StatusNotImplemented)
	}
}

// @Summary		Health Check
// @Description	Returns the health status of the service
// @Tags			Health
// @Success		200	{object}	responses.HealthResponse
// @Router			/api/v1/health [get]
func (h *Handler) HealthCheck(dbClient *mongo.Client) fiber.Handler {
	return func(c *fiber.Ctx) error {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		err := dbClient.Ping(ctx, readpref.Primary())
		if err != nil {
			// status degraded
			health := responses.HealthResponse{
				Status:    responses.HealthStatusDegraded,
				Timestamp: time.Now().UTC().Format(time.RFC3339),
				Errors:    []string{err.Error()},
			}
			return c.JSON(health)
		}
		health := responses.HealthResponse{
			Status:    responses.HealthStatusHealthy,
			Timestamp: time.Now().UTC().Format(time.RFC3339),
			Errors:    []string{},
		}

		return c.JSON(fiber.Map{"result": health})
	}
}

// @Summary		Get Service Configs
// @Description	Returns the service configs
// @Tags			Misc
// @Success		200	{object}	responses.AppConfigsResponse
// @Router			/api/v1/configs [get]
func (h *Handler) GetServiceConfigs() fiber.Handler {
	return func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"result": h.config.GetPublicConfig()})
	}
}

// @Summary		Get Service Version
// @Description	Returns the service version
// @Tags			Misc
// @Success		200	{object}	string
// @Router			/api/v1/version [get]
func (h *Handler) GetServiceVersion() fiber.Handler {
	return func(c *fiber.Ctx) error {
		response := responses.AppVersionResponse{Version: build.Version}
		return c.JSON(fiber.Map{"result": response})
	}
}

func HandleError(c *fiber.Ctx, err error) error {
	if serviceErr, ok := err.(*errors.ServiceError); ok {
		return HandleServiceError(c, serviceErr)
	}
	return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
}

func HandleServiceError(c *fiber.Ctx, err *errors.ServiceError) error {
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
