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
	"context"
	"fmt"
	"log/slog"
	"strconv"
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

const (
	// iDenfy webhook headers
	HeaderIdenfySignature = "Idenfy-Signature"

	// Query parameters
	QueryParamClientID = "client_id"
	QueryParamTwinID   = "twin_id"

	HandlerTimeout = 5 * time.Second
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
// @Failure		403			{object}		object{error=string}
// @Failure		409			{object}		object{error=string}
// @Failure		500			{object}		object{error=string}
// @Failure		503			{object}		object{error=string}
// @Router			/api/v1/token [post]
func (h *Handler) GetOrCreateVerificationToken() fiber.Handler {
	return func(c *fiber.Ctx) error {
		clientID := c.Locals("clientID").(string)
		ctx, cancel := context.WithTimeout(c.Context(), HandlerTimeout)
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
		clientID := c.Locals("clientID").(string)
		ctx, cancel := context.WithTimeout(c.Context(), HandlerTimeout)
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
		clientID := c.Query(QueryParamClientID)
		twinID := c.Query(QueryParamTwinID)

		if clientID == "" && twinID == "" {
			h.logger.Warn("Bad request: missing client_id and twin_id")
			return responses.RespondWithError(c, fiber.StatusBadRequest, fmt.Errorf("either client_id or twin_id must be provided"))
		}
		var verification *models.VerificationOutcome
		var err error
		ctx, cancel := context.WithTimeout(c.Context(), HandlerTimeout)
		defer cancel()
		if clientID != "" {
			verification, err = h.kycService.GetVerificationStatus(ctx, clientID)
		} else {
			twinIDUint64, parseErr := strconv.ParseUint(twinID, 10, 32)
			if parseErr != nil {
				h.logger.Error("Error parsing twinID", "twinID", twinID, "error", parseErr)
				return responses.RespondWithError(c, fiber.StatusBadRequest, fmt.Errorf("invalid twinID"))
			}
			verification, err = h.kycService.GetVerificationStatusByTwinID(ctx, uint32(twinIDUint64))
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
			"headers", &c.Request().Header,
		)
		sigHeader := c.Get(HeaderIdenfySignature)
		if len(sigHeader) < 1 {
			h.logger.Error("Missing signature header", "headers", string(c.Request().Header.Header()))
			return responses.RespondWithError(c, fiber.StatusBadRequest, fmt.Errorf("no signature provided"))
		}
		body := c.Body()
		var result models.Verification
		if err := c.BodyParser(&result); err != nil {
			h.logger.Error("Error decoding verification update", "error", err)
			return responses.RespondWithError(c, fiber.StatusBadRequest, err)
		}
		ctx, cancel := context.WithTimeout(c.Context(), HandlerTimeout)
		defer cancel()
		if err := h.kycService.ProcessVerificationResult(ctx, body, sigHeader, result); err != nil {
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
		h.logger.Debug("Received ID expiration update",
			"body", string(c.Body()),
			"headers", &c.Request().Header,
		)

		// Verify signature
		sigHeader := c.Get(HeaderIdenfySignature)
		if len(sigHeader) < 1 {
			h.logger.Error("Missing signature header", "headers", string(c.Request().Header.Header()))
			return responses.RespondWithError(c, fiber.StatusBadRequest, fmt.Errorf("missing signature header"))
		}
		body := c.Body()
		var notification models.DocExpirationNotification
		if err := c.BodyParser(&notification); err != nil {
			h.logger.Error("Error decoding verification update", "error", err)
			return responses.RespondWithError(c, fiber.StatusBadRequest, fmt.Errorf("invalid request body"))
		}
		ctx, cancel := context.WithTimeout(c.Context(), HandlerTimeout)
		defer cancel()
		if err := h.kycService.ProcessDocExpirationNotification(ctx, body, sigHeader, notification); err != nil {
			return HandleError(c, err)
		}

		return c.SendStatus(fiber.StatusOK)
	}
}

// @Summary		Health Check
// @Description	Returns the health status of the service
// @Tags			Health
// @Success		200	{object}	object{result=responses.HealthResponse}
// @Router			/api/v1/health [get]
func (h *Handler) HealthCheck(dbClient *mongo.Client) fiber.Handler {
	return func(c *fiber.Ctx) error {
		ctx, cancel := context.WithTimeout(c.Context(), HandlerTimeout)
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

// CreateSponsorship creates a new sponsorship between a KYC-verified sponsor and a sponsee
// @Summary Create a new sponsorship
// @Description Creates a new sponsorship where a KYC-verified twin sponsors another twin. Both sponsor and sponsee must authenticate.
// @Tags Sponsorships
// @Accept json
// @Produce json
// @Param X-Client-ID header string true "TFChain SS58Address of the sponsor" minlength(48) maxlength(48)
// @Param X-Challenge header string true "hex-encoded message `{api-domain}:{timestamp}`"
// @Param X-Signature header string true "hex-encoded sr25519|ed25519 signature of the sponsor" minlength(128) maxlength(128)
// @Param X-Sponsee-ID header string true "TFChain SS58Address of the sponsee" minlength(48) maxlength(48)
// @Param X-Sponsee-Challenge header string true "hex-encoded message `{api-domain}:{timestamp}`"
// @Param X-Sponsee-Signature header string true "hex-encoded sr25519|ed25519 signature of the sponsee" minlength(128) maxlength(128)
// @Success 201 {object} responses.Response{data=models.Sponsorship} "Sponsorship created successfully"
// @Failure 400 {object} responses.ErrorResponse "Bad request"
// @Failure 401 {object} responses.ErrorResponse "Unauthorized"
// @Failure 403 {object} responses.ErrorResponse "Forbidden"
// @Failure 404 {object} responses.ErrorResponse "Not Found"
// @Failure 409 {object} responses.ErrorResponse "Conflict"
// @Failure 500 {object} responses.ErrorResponse "Internal Server Error"
// @Router /api/v1/sponsorships [post]
func (h *Handler) CreateSponsorship() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Get the Sponsee twin ID from the AuthMiddleware (set by SponseeAuthMiddleware)
		sponseeClientID, ok := c.Locals("sponseeID").(string)
		if !ok || sponseeClientID == "" {
			h.logger.Error("missing or invalid sponsee client ID in context")
			return responses.RespondWithError(c, fiber.StatusInternalServerError, fmt.Errorf("missing sponsee authentication"))
		}

		// Get the sponsor's client ID from the AuthMiddleware
		sponsorClientID, ok := c.Locals("clientID").(string)
		if !ok || sponsorClientID == "" {
			h.logger.Error("missing or invalid sponsor client ID in context")
			return responses.RespondWithError(c, fiber.StatusInternalServerError,
				fmt.Errorf("missing sponsor authentication"))
		}

		sponsorship, err := h.kycService.CreateSponsorship(c.Context(), sponsorClientID, sponseeClientID)
		if err != nil {
			return HandleError(c, err)
		}
		return responses.RespondWithData(c, fiber.StatusCreated, sponsorship)
	}
}

// GetSponsorships retrieves a list of sponsorships with optional filtering and pagination
// @Summary List sponsorships with optional filtering
// @Description Returns a paginated list of sponsorships. If no filter is provided, returns all sponsorships with pagination.
// @Tags Sponsorships
// @Produce json
// @Param sponsor_twin_id query int false "Filter by sponsor twin ID (mutually exclusive with sponsee_twin_id)"
// @Param sponsee_twin_id query int false "Filter by sponsee twin ID (mutually exclusive with sponsor_twin_id)"
// @Param limit query int false "Maximum number of results to return (default: 50, max: 100)"
// @Param offset query int false "Number of results to skip for pagination (default: 0)"
// @Success 200 {object} responses.PaginatedResponse{data=[]models.Sponsorship} "Paginated list of sponsorships"
// @Failure 400 {object} responses.ErrorResponse "Bad request"
// @Failure 404 {object} responses.ErrorResponse "Not Found"
// @Failure 500 {object} responses.ErrorResponse "Internal Server Error"
// @Router /api/v1/sponsorships [get]
func (h *Handler) GetSponsorships() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Parse query parameters
		sponsorTwinID, err := getUint32Param(c, "sponsor_twin_id")
		if err != nil {
			h.logger.Error("invalid sponsor_twin_id parameter", "error", err)
			return responses.RespondWithError(c, fiber.StatusBadRequest, fmt.Errorf("invalid sponsor_twin_id parameter"))
		}

		sponseeTwinID, err := getUint32Param(c, "sponsee_twin_id")
		if err != nil {
			h.logger.Error("invalid sponsee_twin_id parameter", "error", err)
			return responses.RespondWithError(c, fiber.StatusBadRequest, fmt.Errorf("invalid sponsee_twin_id parameter"))
		}

		// Parse and validate pagination parameters with default limit of 50, default offset of 0 and max limit of 100
		limit, err := strconv.ParseInt(c.Query("limit", "50"), 10, 64)
		if err != nil || limit < 1 {
			return responses.RespondWithError(c, fiber.StatusBadRequest, fmt.Errorf("invalid limit parameter"))
		}
		if limit > 100 {
			return responses.RespondWithError(c, fiber.StatusBadRequest, fmt.Errorf("limit parameter too large"))
		}

		offset, err := strconv.ParseInt(c.Query("offset", "0"), 10, 64)
		if err != nil || offset < 0 {
			return responses.RespondWithError(c, fiber.StatusBadRequest, fmt.Errorf("invalid offset parameter"))
		}

		// If no filters provided, return all sponsorships with pagination
		if sponsorTwinID == 0 && sponseeTwinID == 0 {
			sponsorships, total, err := h.kycService.ListAllSponsorships(c.Context(), limit, offset)
			if err != nil {
				return HandleError(c, err)
			}

			// Return empty array instead of null for consistency
			if sponsorships == nil {
				sponsorships = []*models.Sponsorship{}
			}

			return responses.RespondWithPagination(c, fiber.StatusOK, sponsorships, total, limit, offset)
		}

		// Validate that only one filter is provided when filtering
		if sponsorTwinID > 0 && sponseeTwinID > 0 {
			h.logger.Error("only one of sponsor_twin_id or sponsee_twin_id can be provided")
			return responses.RespondWithError(c, fiber.StatusBadRequest,
				fmt.Errorf("only one of sponsor_twin_id or sponsee_twin_id can be provided"))
		}

		var sponsorships []*models.Sponsorship
		var total int64

		// Get sponsorships based on the provided filter
		if sponsorTwinID > 0 {
			sponsorships, total, err = h.kycService.GetSponsorshipsBySponsor(c.Context(), uint32(sponsorTwinID), limit, offset)
		} else if sponseeTwinID > 0 {
			var sponsorship *models.Sponsorship
			sponsorship, err = h.kycService.GetSponsorshipBySponsee(c.Context(), uint32(sponseeTwinID))
			if sponsorship != nil {
				sponsorships = []*models.Sponsorship{sponsorship}
				total = 1
			} else {
				sponsorships = []*models.Sponsorship{}
				total = 0
			}
		}

		if err != nil {
			return HandleError(c, err)
		}

		// Return empty array instead of null for consistency
		if sponsorships == nil {
			sponsorships = []*models.Sponsorship{}
		}

		return responses.RespondWithPagination(c, fiber.StatusOK, sponsorships, total, limit, offset)
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
	case errors.ErrorTypeForbidden:
		return fiber.StatusForbidden
	default:
		return fiber.StatusInternalServerError
	}
}

// getUint32Param parses a query parameter as uint32
func getUint32Param(c *fiber.Ctx, param string) (uint32, error) {
	value := c.Query(param)
	if value == "" {
		return 0, nil
	}

	intValue, err := strconv.ParseUint(value, 10, 32)
	if err != nil {
		return 0, err
	}

	return uint32(intValue), nil
}
