/*
Package server contains the HTTP server for the application.
This layer is responsible for initializing the server and its dependencies. in more details:
- setting up the middleware
- setting up the database
- setting up the repositories
- setting up the services
- setting up the routes
*/
package server

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/helmet"
	"github.com/gofiber/fiber/v2/middleware/limiter"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/storage/mongodb"
	"github.com/gofiber/swagger"
	_ "github.com/threefoldtech/tf-kyc-verifier/api/docs"
	"github.com/threefoldtech/tf-kyc-verifier/internal/clients/idenfy"
	"github.com/threefoldtech/tf-kyc-verifier/internal/clients/substrate"
	"github.com/threefoldtech/tf-kyc-verifier/internal/config"
	"github.com/threefoldtech/tf-kyc-verifier/internal/handlers"
	"github.com/threefoldtech/tf-kyc-verifier/internal/logger"
	"github.com/threefoldtech/tf-kyc-verifier/internal/middleware"
	"github.com/threefoldtech/tf-kyc-verifier/internal/repository"
	"github.com/threefoldtech/tf-kyc-verifier/internal/services"
	"go.mongodb.org/mongo-driver/mongo"
)

// Server represents the HTTP server and its dependencies
type Server struct {
	app    *fiber.App
	config *config.Config
	logger logger.Logger
}

// New creates a new server instance with the given configuration and options
func New(config *config.Config, srvLogger logger.Logger) (*Server, error) {
	// Create base context for initialization
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Initialize server with base configuration
	server := &Server{
		config: config,
		logger: srvLogger,
	}

	// Initialize Fiber app with base configuration
	server.app = fiber.New(fiber.Config{
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  20 * time.Second,
		BodyLimit:    512 * 1024, // 512KB
	})

	// Initialize core components
	if err := server.initializeCore(ctx); err != nil {
		return nil, fmt.Errorf("initializing core components: %w", err)
	}

	return server, nil
}

// initializeCore sets up the core components of the server
func (s *Server) initializeCore(ctx context.Context) error {
	// Setup middleware
	if err := s.setupMiddleware(); err != nil {
		return fmt.Errorf("setting up middleware: %w", err)
	}

	// Setup database
	dbClient, db, err := s.setupDatabase(ctx)
	if err != nil {
		return fmt.Errorf("setting up database: %w", err)
	}

	// Setup repositories
	repos, err := s.setupRepositories(ctx, db)
	if err != nil {
		return fmt.Errorf("setting up repositories: %w", err)
	}

	// Setup services
	service, err := s.setupServices(repos)
	if err != nil {
		return fmt.Errorf("setting up services: %w", err)
	}

	// Setup routes
	if err := s.setupRoutes(service, dbClient); err != nil {
		return fmt.Errorf("setting up routes: %w", err)
	}

	return nil
}

func (s *Server) setupMiddleware() error {
	s.logger.Debug("Setting up middleware", nil)

	// Setup rate limiter stores
	ipLimiterStore := mongodb.New(mongodb.Config{
		ConnectionURI: s.config.MongoDB.URI,
		Database:      s.config.MongoDB.DatabaseName,
		Collection:    "ip_limit",
		Reset:         false,
	})

	idLimiterStore := mongodb.New(mongodb.Config{
		ConnectionURI: s.config.MongoDB.URI,
		Database:      s.config.MongoDB.DatabaseName,
		Collection:    "id_limit",
		Reset:         false,
	})

	// Configure rate limiters
	ipLimiterConfig := limiter.Config{
		Max:        int(s.config.IPLimiter.MaxTokenRequests),
		Expiration: time.Duration(s.config.IPLimiter.TokenExpiration) * time.Minute,
		Storage:    ipLimiterStore,
		KeyGenerator: func(c *fiber.Ctx) string {
			return extractIPFromRequest(c)
		},
		Next: func(c *fiber.Ctx) bool {
			return extractIPFromRequest(c) == "127.0.0.1"
		},
		SkipFailedRequests: true,
	}

	idLimiterConfig := limiter.Config{
		Max:        int(s.config.IDLimiter.MaxTokenRequests),
		Expiration: time.Duration(s.config.IDLimiter.TokenExpiration) * time.Minute,
		Storage:    idLimiterStore,
		KeyGenerator: func(c *fiber.Ctx) string {
			return c.Get("X-Client-ID")
		},
		SkipFailedRequests: true,
	}

	// Apply middleware
	s.app.Use(middleware.NewLoggingMiddleware(s.logger))
	s.app.Use(middleware.CORS())
	s.app.Use(recover.New(recover.Config{
		EnableStackTrace: true,
	}))
	s.app.Use(helmet.New())

	if s.config.IPLimiter.MaxTokenRequests > 0 {
		s.app.Use("/api/v1/token", limiter.New(ipLimiterConfig))
	}
	if s.config.IDLimiter.MaxTokenRequests > 0 {
		s.app.Use("/api/v1/token", limiter.New(idLimiterConfig))
	}

	return nil
}

func (s *Server) setupDatabase(ctx context.Context) (*mongo.Client, *mongo.Database, error) {
	s.logger.Debug("Connecting to database", nil)

	client, err := repository.ConnectToMongoDB(ctx, s.config.MongoDB.URI)
	if err != nil {
		return nil, nil, fmt.Errorf("setting up database: %w", err)
	}

	return client, client.Database(s.config.MongoDB.DatabaseName), nil
}

type repositories struct {
	token        repository.TokenRepository
	verification repository.VerificationRepository
}

func (s *Server) setupRepositories(ctx context.Context, db *mongo.Database) (*repositories, error) {
	s.logger.Debug("Setting up repositories", nil)

	return &repositories{
		token:        repository.NewMongoTokenRepository(ctx, db, s.logger),
		verification: repository.NewMongoVerificationRepository(ctx, db, s.logger),
	}, nil
}

func (s *Server) setupServices(repos *repositories) (services.KYCService, error) {
	s.logger.Debug("Setting up services", nil)

	idenfyClient := idenfy.New(&s.config.Idenfy, s.logger)

	substrateClient, err := substrate.New(&s.config.TFChain, s.logger)
	if err != nil {
		return nil, fmt.Errorf("initializing substrate client: %w", err)
	}
	kycService, err := services.NewKYCService(
		repos.verification,
		repos.token,
		idenfyClient,
		substrateClient,
		s.config,
		s.logger,
	)
	if err != nil {
		return nil, err
	}
	return kycService, nil
}

func (s *Server) setupRoutes(kycService services.KYCService, mongoCl *mongo.Client) error {
	s.logger.Debug("Setting up routes", nil)

	handler := handlers.NewHandler(kycService, s.config, s.logger)

	// API routes
	v1 := s.app.Group("/api/v1")
	v1.Post("/token", middleware.AuthMiddleware(s.config.Challenge), handler.GetOrCreateVerificationToken())
	v1.Get("/data", middleware.AuthMiddleware(s.config.Challenge), handler.GetVerificationData())
	v1.Get("/status", handler.GetVerificationStatus())
	v1.Get("/health", handler.HealthCheck(mongoCl))
	v1.Get("/configs", handler.GetServiceConfigs())
	v1.Get("/version", handler.GetServiceVersion())

	// Webhook routes
	webhooks := s.app.Group("/webhooks/idenfy")
	webhooks.Post("/verification-update", handler.ProcessVerificationResult())
	webhooks.Post("/id-expiration", handler.ProcessDocExpirationNotification())

	// Documentation
	s.app.Get("/docs/*", swagger.HandlerDefault)

	return nil
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

func (s *Server) Run() error {
	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
		<-sigChan
		// Graceful shutdown
		s.logger.Info("Shutting down server...", nil)
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		if err := s.app.ShutdownWithContext(ctx); err != nil {
			s.logger.Error("Server forced to shutdown:", logger.Fields{"error": err})
		}
	}()

	// Start server
	if err := s.app.Listen(":" + s.config.Server.Port); err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("starting server: %w", err)
	}
	return nil
}
