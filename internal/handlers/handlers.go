/*
Package handlers contains the handlers for the API.
This layer is responsible for handling the requests and responses, in more details:
- validating the requests
- formatting the responses
- handling the errors
- delegating the requests to the services
*/
package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/readpref"

	"github.com/threefoldtech/tf-kyc-verifier/internal/build"
	"github.com/threefoldtech/tf-kyc-verifier/internal/config"
	"github.com/threefoldtech/tf-kyc-verifier/internal/errors"
	"github.com/threefoldtech/tf-kyc-verifier/internal/models"
	"github.com/threefoldtech/tf-kyc-verifier/internal/responses"
	"github.com/threefoldtech/tf-kyc-verifier/internal/services"
)

type Handler struct {
	kycService *services.KYCService
	config     *config.Config
	logger     *slog.Logger
}

//	@title			TFGrid KYC API
//	@version		0.2.0
//	@description	This is a KYC service for TFGrid.
//	@termsOfService	http://swagger.io/terms/

// @contact.name	threefold.io
// @contact.url		https://threefold.io
// @contact.email	info@threefold.io
// @BasePath		/
func NewHandler(kycService *services.KYCService, config *config.Config, logger *slog.Logger) *Handler {
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
// @Success		200			{object}		object{result=responses.TokenResponse} "Existing token retrieved"
// @Success		201			{object}		object{result=responses.TokenResponse} "New token created"
// @Failure		400			{object}		object{error=string}
// @Failure		401			{object}		object{error=string}
// @Failure		402			{object}		object{error=string}
// @Failure		409			{object}		object{error=string}
// @Failure		500			{object}		object{error=string}
// @Failure		503			{object}		object{error=string}
// @Router			/api/v1/token [post]
func (h *Handler) GetOrCreateVerificationToken() fiber.Handler {
	return func(c *fiber.Ctx) error {
		clientID := c.Get("X-Client-ID")
		ctx, cancel := context.WithTimeout(c.Context(), 5*time.Second)
		defer cancel()
		token, isNewToken, err := h.kycService.GetOrCreateVerificationToken(ctx, clientID)
		if err != nil {
			return HandleError(c, err)
		}
		response := responses.NewTokenResponseWithStatus(token, isNewToken)
		if isNewToken {
			return responses.RespondWithData(c, fiber.StatusCreated, response)
		}
		return responses.RespondWithData(c, fiber.StatusOK, response)
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
// @Success		200			{object}		object{result=responses.VerificationDataResponse}
// @Failure		400			{object}		object{error=string}
// @Failure		401			{object}		object{error=string}
// @Failure		404			{object}		object{error=string}
// @Failure		500			{object}		object{error=string}
// @Router			/api/v1/data [get]
func (h *Handler) GetVerificationData() fiber.Handler {
	return func(c *fiber.Ctx) error {
		clientID := c.Get("X-Client-ID")
		ctx, cancel := context.WithTimeout(c.Context(), 5*time.Second)
		defer cancel()
		verification, err := h.kycService.GetVerificationData(ctx, clientID)
		if err != nil {
			return HandleError(c, err)
		}
		if verification == nil {
			return responses.RespondWithError(c, fiber.StatusNotFound, fmt.Errorf("verification not found for client"))
		}
		response := responses.NewVerificationDataResponse(verification)
		return responses.RespondWithData(c, fiber.StatusOK, response)
	}
}

// @Summary		Get Verification Status
// @Description	Returns the verification status for a client
// @Tags			Verification
// @Accept			json
// @Produce		json
// @Param			client_id	query		string	false	"TFChain SS58Address"								minlength(48)	maxlength(48)
// @Param			twin_id		query		string	false	"Twin ID"											minlength(1)
// @Success		200			{object}		object{result=responses.VerificationStatusResponse}
// @Failure		400			{object}		object{error=string}
// @Failure		404			{object}		object{error=string}
// @Failure		500			{object}		object{error=string}
// @Failure		503			{object}		object{error=string}
// @Router			/api/v1/status [get]
func (h *Handler) GetVerificationStatus() fiber.Handler {
	return func(c *fiber.Ctx) error {
		clientID := c.Query("client_id")
		twinID := c.Query("twin_id")

		if clientID == "" && twinID == "" {
			h.logger.Warn("Bad request: missing client_id and twin_id")
			return responses.RespondWithError(c, fiber.StatusBadRequest, fmt.Errorf("either client_id or twin_id must be provided"))
		}
		var verification *models.VerificationOutcome
		var err error
		ctx, cancel := context.WithTimeout(c.Context(), 5*time.Second)
		defer cancel()
		if clientID != "" {
			verification, err = h.kycService.GetVerificationStatus(ctx, clientID)
		} else {
			verification, err = h.kycService.GetVerificationStatusByTwinID(ctx, twinID)
		}
		if err != nil {
			h.logger.Error("Failed to get verification status", "clientID", clientID, "twinID", twinID, "error", err)
			return HandleError(c, err)
		}
		if verification == nil {
			h.logger.Info("Verification not found", "clientID", clientID, "twinID", twinID)
			return responses.RespondWithError(c, fiber.StatusNotFound, fmt.Errorf("verification not found"))
		}
		response := responses.NewVerificationStatusResponse(verification)
		return responses.RespondWithData(c, fiber.StatusOK, response)
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
		h.logger.Debug("Received verification update",
			"body", string(c.Body()),
			"headers", &c.Request().Header,
		)
		sigHeader := c.Get("Idenfy-Signature")
		if len(sigHeader) < 1 {
			return responses.RespondWithError(c, fiber.StatusBadRequest, fmt.Errorf("no signature provided"))
		}
		body := c.Body()
		var result models.Verification
		decoder := json.NewDecoder(bytes.NewReader(body))
		err := decoder.Decode(&result)
		if err != nil {
			h.logger.Error("Error decoding verification update", "error", err)
			return responses.RespondWithError(c, fiber.StatusBadRequest, err)
		}
		h.logger.Debug("Verification update after decoding", "result", result)
		ctx, cancel := context.WithTimeout(c.Context(), 5*time.Second)
		defer cancel()
		err = h.kycService.ProcessVerificationResult(ctx, body, sigHeader, result)
		if err != nil {
			return HandleError(c, err)
		}
		return responses.RespondWithData(c, fiber.StatusOK, nil)
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
// @Success		200	{object}	object{result=responses.HealthResponse}
// @Router			/api/v1/health [get]
func (h *Handler) HealthCheck(dbClient *mongo.Client) fiber.Handler {
	return func(c *fiber.Ctx) error {
		ctx, cancel := context.WithTimeout(c.Context(), 5*time.Second)
		defer cancel()
		err := dbClient.Ping(ctx, readpref.Primary())
		if err != nil {
			// status degraded
			health := responses.HealthResponse{
				Status:    responses.HealthStatusDegraded,
				Timestamp: time.Now().UTC().Format(time.RFC3339),
				Errors:    []string{err.Error()},
			}
			return responses.RespondWithData(c, fiber.StatusOK, health)
		}
		health := responses.HealthResponse{
			Status:    responses.HealthStatusHealthy,
			Timestamp: time.Now().UTC().Format(time.RFC3339),
			Errors:    []string{},
		}

		return responses.RespondWithData(c, fiber.StatusOK, health)
	}
}

// @Summary		Get Service Configs
// @Description	Returns the service configs
// @Tags			Misc
// @Success		200	{object}	object{result=responses.AppConfigsResponse}
// @Router			/api/v1/configs [get]
func (h *Handler) GetServiceConfigs() fiber.Handler {
	return func(c *fiber.Ctx) error {
		return responses.RespondWithData(c, fiber.StatusOK, h.config.GetPublicConfig())
	}
}

// @Summary		Get Service Version
// @Description	Returns the service version
// @Tags			Misc
// @Success		200	{object}	object{result=responses.AppVersionResponse}
// @Router			/api/v1/version [get]
func (h *Handler) GetServiceVersion() fiber.Handler {
	return func(c *fiber.Ctx) error {
		response := responses.AppVersionResponse{Version: build.Version}
		return responses.RespondWithData(c, fiber.StatusOK, response)
	}
}

func HandleError(c *fiber.Ctx, err error) error {
	if serviceErr, ok := err.(*errors.ServiceError); ok {
		return HandleServiceError(c, serviceErr)
	}
	return responses.RespondWithError(c, fiber.StatusInternalServerError, err)
}

func HandleServiceError(c *fiber.Ctx, err *errors.ServiceError) error {
	statusCode := getStatusCode(err.Type)
	return responses.RespondWithError(c, statusCode, err)
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
		return fiber.StatusServiceUnavailable
	case errors.ErrorTypeNotSufficientBalance:
		return fiber.StatusPaymentRequired
	default:
		return fiber.StatusInternalServerError
	}
}
