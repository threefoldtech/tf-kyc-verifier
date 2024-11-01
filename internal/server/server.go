package server

import (
	"context"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	_ "example.com/tfgrid-kyc-service/api/docs"
	"example.com/tfgrid-kyc-service/internal/clients/idenfy"
	"example.com/tfgrid-kyc-service/internal/clients/substrate"
	"example.com/tfgrid-kyc-service/internal/configs"
	"example.com/tfgrid-kyc-service/internal/handlers"
	"example.com/tfgrid-kyc-service/internal/logger"
	"example.com/tfgrid-kyc-service/internal/middleware"
	"example.com/tfgrid-kyc-service/internal/repository"
	"example.com/tfgrid-kyc-service/internal/services"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/helmet"
	"github.com/gofiber/fiber/v2/middleware/limiter"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/storage/mongodb"
	"github.com/gofiber/swagger"
)

// implement server struct that have fiber app and config
type Server struct {
	app    *fiber.App
	config *configs.Config
	logger logger.Logger
}

func New(config *configs.Config, logger logger.Logger) *Server {
	// debug log
	app := fiber.New(fiber.Config{
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
		BodyLimit:    512 * 1024, // 512KB
	})
	// Setup Limter Config and store
	ipLimiterstore := mongodb.New(mongodb.Config{
		ConnectionURI: config.MongoDB.URI,
		Database:      config.MongoDB.DatabaseName,
		Collection:    "ip_limit",
		Reset:         false,
	})
	ipLimiterConfig := limiter.Config{
		Max:                    config.IPLimiter.MaxTokenRequests,
		Expiration:             time.Duration(config.IPLimiter.TokenExpiration) * time.Minute,
		SkipFailedRequests:     true,
		SkipSuccessfulRequests: false,
		Storage:                ipLimiterstore,
		// skip the limiter for localhost
		Next: func(c *fiber.Ctx) bool {
			// skip the limiter if the keyGenerator returns "127.0.0.1"
			return extractIPFromRequest(c) == "127.0.0.1"
		},
		KeyGenerator: func(c *fiber.Ctx) string {
			return extractIPFromRequest(c)
		},
	}
	idLimiterStore := mongodb.New(mongodb.Config{
		ConnectionURI: config.MongoDB.URI,
		Database:      config.MongoDB.DatabaseName,
		Collection:    "id_limit",
		Reset:         false,
	})

	idLimiterConfig := limiter.Config{
		Max:                    config.IDLimiter.MaxTokenRequests,
		Expiration:             time.Duration(config.IDLimiter.TokenExpiration) * time.Minute,
		SkipFailedRequests:     true,
		SkipSuccessfulRequests: false,
		Storage:                idLimiterStore,
		// Use client id as key to limit the number of requests per client
		KeyGenerator: func(c *fiber.Ctx) string {
			return c.Get("X-Client-ID")
		},
	}

	// Global middlewares
	app.Use(middleware.NewLoggingMiddleware(logger))
	app.Use(middleware.CORS())
	recoverConfig := recover.ConfigDefault
	recoverConfig.EnableStackTrace = true
	app.Use(recover.New(recoverConfig))
	app.Use(helmet.New())

	// Database connection
	db, err := repository.ConnectToMongoDB(config.MongoDB.URI)
	if err != nil {
		logger.Fatal("Failed to connect to MongoDB", map[string]interface{}{"error": err})
	}
	database := db.Database(config.MongoDB.DatabaseName)

	// Initialize repositories
	tokenRepo := repository.NewMongoTokenRepository(database, logger)
	verificationRepo := repository.NewMongoVerificationRepository(database, logger)

	// Initialize services
	idenfyClient := idenfy.New(&config.Idenfy, logger)

	substrateClient, err := substrate.New(&config.TFChain, logger)
	if err != nil {
		logger.Fatal("Failed to initialize substrate client", map[string]interface{}{"error": err})
	}
	kycService := services.NewKYCService(verificationRepo, tokenRepo, idenfyClient, substrateClient, config, logger)

	// Initialize handler
	handler := handlers.NewHandler(kycService, config, logger)

	// Routes
	app.Get("/docs/*", swagger.HandlerDefault)

	v1 := app.Group("/api/v1")
	v1.Post("/token", middleware.AuthMiddleware(config.Challenge), limiter.New(idLimiterConfig), limiter.New(ipLimiterConfig), handler.GetorCreateVerificationToken())
	v1.Get("/data", middleware.AuthMiddleware(config.Challenge), handler.GetVerificationData())
	// status route accepts either client_id or twin_id as query parameters
	v1.Get("/status", handler.GetVerificationStatus())
	v1.Get("/health", handler.HealthCheck(db))
	v1.Get("/configs", handler.GetServiceConfigs())
	v1.Get("/version", handler.GetServiceVersion())
	// Webhook routes
	webhooks := app.Group("/webhooks/idenfy")
	webhooks.Post("/verification-update", handler.ProcessVerificationResult())
	webhooks.Post("/id-expiration", handler.ProcessDocExpirationNotification())

	return &Server{app: app, config: config, logger: logger}
}

func extractIPFromRequest(c *fiber.Ctx) string {
	// Check for X-Forwarded-For header
	if ip := c.Get("X-Forwarded-For"); ip != "" {
		ips := strings.Split(ip, ",")
		if len(ips) > 0 {

			for _, ip := range ips {
				// return the first non-private ip in the list
				if net.ParseIP(strings.TrimSpace(ip)) != nil && !net.ParseIP(strings.TrimSpace(ip)).IsPrivate() {
					return strings.TrimSpace(ip)
				}
			}
		}
	}
	// Check for X-Real-IP header if not a private IP
	if ip := c.Get("X-Real-IP"); ip != "" {
		if net.ParseIP(strings.TrimSpace(ip)) != nil && !net.ParseIP(strings.TrimSpace(ip)).IsPrivate() {
			return strings.TrimSpace(ip)
		}
	}
	// Fall back to RemoteIP() if no proxy headers are present
	ip := c.IP()
	if parsedIP := net.ParseIP(ip); parsedIP != nil {
		if !parsedIP.IsPrivate() {
			return ip
		}
	}
	// If we still have a private IP, return a default value that will be skipped by the limiter
	return "127.0.0.1"
}

func (s *Server) Start() {
	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
		<-sigChan
		// Graceful shutdown
		s.logger.Info("Shutting down server...", map[string]interface{}{})
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		if err := s.app.ShutdownWithContext(ctx); err != nil {
			s.logger.Error("Server forced to shutdown:", map[string]interface{}{"error": err})
		}
	}()

	// Start server
	if err := s.app.Listen(":" + s.config.Server.Port); err != nil && err != http.ErrServerClosed {
		s.logger.Fatal("Server startup failed", map[string]interface{}{"error": err})
	}
}
