package server

import (
	"fmt"
	"log"
	"net"
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
}

func New(config *configs.Config) *Server {
	// debug log
	app := fiber.New()

	// Setup Limter Config and store
	ipLimiterstore := mongodb.New(mongodb.Config{
		ConnectionURI: config.MongoDB.URI,
		Database:      config.MongoDB.DatabaseName,
		Collection:    "ip_limit",
		Reset:         false,
	})
	ipLimiterConfig := limiter.Config{ // TODO: use configurable parameters, also check if it works well after passing the request through an SSL gateway
		Max:                    config.IPLimiter.MaxTokenRequests,
		Expiration:             time.Duration(config.IPLimiter.TokenExpiration) * time.Minute,
		SkipFailedRequests:     false,
		SkipSuccessfulRequests: false,
		Store:                  ipLimiterstore,
		// skip the limiter for localhost
		Next: func(c *fiber.Ctx) bool {
			return c.IP() == "127.0.0.1"
		},
		KeyGenerator: func(c *fiber.Ctx) string {
			// Check for X-Forwarded-For header
			if ip := c.Get("X-Forwarded-For"); ip != "" {
				ips := strings.Split(ip, ",")
				if len(ips) > 0 {
					// return the first non-private ip in the list
					for _, ip := range ips {
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
		},
	}
	idLimiterStore := mongodb.New(mongodb.Config{
		ConnectionURI: config.MongoDB.URI,
		Database:      config.MongoDB.DatabaseName,
		Collection:    "id_limit",
		Reset:         false,
	})

	idLimiterConfig := limiter.Config{ // TODO: use configurable parameters
		Max:                    config.IDLimiter.MaxTokenRequests,
		Expiration:             time.Duration(config.IDLimiter.TokenExpiration) * time.Minute,
		SkipFailedRequests:     false,
		SkipSuccessfulRequests: false,
		Store:                  idLimiterStore,
		// Use client id as key to limit the number of requests per client
		KeyGenerator: func(c *fiber.Ctx) string {
			return c.Get("X-Client-ID")
		},
	}
	// print limtters config
	fmt.Printf("IP Limiter Config: %+v\n", ipLimiterConfig)
	fmt.Printf("ID Limiter Config: %+v\n", idLimiterConfig)

	// Global middlewares
	app.Use(middleware.Logger())
	app.Use(middleware.CORS())
	app.Use(recover.New())
	app.Use(helmet.New())

	// Database connection
	db, err := repository.ConnectToMongoDB(config.MongoDB.URI)
	if err != nil {
		log.Fatalf("Failed to connect to MongoDB: %v", err)
	}
	database := db.Database(config.MongoDB.DatabaseName)

	// Initialize repositories
	tokenRepo := repository.NewMongoTokenRepository(database)
	verificationRepo := repository.NewMongoVerificationRepository(database)

	// Initialize services
	idenfyClient := idenfy.New(config.Idenfy)

	if err != nil {
		log.Fatalf("Failed to initialize idenfy client: %v", err)
	}

	substrateClient, err := substrate.New(config.TFChain)
	if err != nil {
		log.Fatalf("Failed to initialize substrate client: %v", err)
	}
	kycService := services.NewKYCService(verificationRepo, tokenRepo, idenfyClient, substrateClient, &config.Verification)

	// Initialize handler
	handler := handlers.NewHandler(kycService)

	// Routes
	app.Get("/docs/*", swagger.HandlerDefault)

	v1 := app.Group("/api/v1")
	v1.Post("/token", middleware.AuthMiddleware(config.ChallengeWindow), limiter.New(idLimiterConfig), limiter.New(ipLimiterConfig), handler.GetorCreateVerificationToken())
	v1.Get("/data", middleware.AuthMiddleware(config.ChallengeWindow), handler.GetVerificationData())
	// status route accepts either client_id or twin_id as query parameters
	v1.Get("/status", handler.GetVerificationStatus())

	// Webhook routes
	webhooks := app.Group("/webhooks/idenfy")
	webhooks.Post("/verification-update", handler.ProcessVerificationResult())
	webhooks.Post("/id-expiration", handler.ProcessDocExpirationNotification())

	return &Server{app: app, config: config}
}

func (s *Server) Start() {
	// Start server
	go func() {
		if err := s.app.Listen(":" + s.config.Server.Port); err != nil {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	if err := s.app.Shutdown(); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exiting")
}
