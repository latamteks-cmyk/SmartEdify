package main

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/smartedify/auth-service/internal/config"
	"github.com/smartedify/auth-service/internal/database"
	"github.com/smartedify/auth-service/internal/handlers"
	"github.com/smartedify/auth-service/internal/hsm"
	"github.com/smartedify/auth-service/internal/jwt"
	"github.com/smartedify/auth-service/internal/middleware"
	"github.com/smartedify/auth-service/internal/repository"
	"github.com/smartedify/auth-service/internal/server"
	"github.com/smartedify/auth-service/internal/service"
	"github.com/smartedify/auth-service/internal/tracing"
	"github.com/smartedify/auth-service/internal/whatsapp"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/helmet"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/redis/go-redis/v9"
	"go.opentelemetry.io/otel/trace"
)

type Application struct {
	Config   *config.Config
	DB       *sql.DB
	Redis    *redis.Client
	Logger   *slog.Logger
	Tracer   trace.Tracer
	Server   *fiber.App
}

func main() {
	// Initialize logger
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		logger.Error("Failed to load configuration", "error", err)
		os.Exit(1)
	}

	// Initialize tracing
	tracerProvider, err := tracing.InitTracer(tracing.TracingConfig{
		ServiceName:    "auth-service",
		ServiceVersion: "1.0.0",
		JaegerEndpoint: cfg.Monitoring.JaegerEndpoint,
		Environment:    cfg.Environment,
	})
	if err != nil {
		logger.Error("Failed to initialize tracer", "error", err)
		os.Exit(1)
	}
	defer func() {
		if err := tracerProvider.Shutdown(context.Background()); err != nil {
			logger.Error("Failed to shutdown tracer", "error", err)
		}
	}()

	// Initialize database
	db, err := database.Connect(cfg.Database)
	if err != nil {
		logger.Error("Failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	// Run database migrations
	if err := database.Migrate(cfg.Database.URL); err != nil {
		logger.Error("Failed to run database migrations", "error", err)
		os.Exit(1)
	}

	// Initialize Redis
	redisClient := redis.NewClient(&redis.Options{
		Addr:     cfg.Redis.Address,
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	})
	defer redisClient.Close()

	// Test Redis connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := redisClient.Ping(ctx).Err(); err != nil {
		logger.Error("Failed to connect to Redis", "error", err)
		os.Exit(1)
	}

	// Create application instance
	app := &Application{
		Config: cfg,
		DB:     db,
		Redis:  redisClient,
		Logger: logger,
		Tracer: tracerProvider.Tracer("auth-service"),
	}

	// Initialize Fiber app
	app.Server = fiber.New(fiber.Config{
		ErrorHandler: server.ErrorHandler,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  120 * time.Second,
	})

	// Setup middleware
	app.setupMiddleware()

	// Setup routes
	app.setupRoutes()

	// Start server
	go func() {
		logger.Info("Starting server", "port", cfg.Server.Port)
		if err := app.Server.Listen(fmt.Sprintf(":%d", cfg.Server.Port)); err != nil {
			logger.Error("Server failed to start", "error", err)
			os.Exit(1)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down server...")

	// Graceful shutdown with timeout
	ctx, cancel = context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := app.Server.ShutdownWithContext(ctx); err != nil {
		logger.Error("Server forced to shutdown", "error", err)
	}

	logger.Info("Server exited")
}

func (app *Application) setupMiddleware() {
	// Recovery middleware
	app.Server.Use(recover.New())

	// Logger middleware
	app.Server.Use(logger.New(logger.Config{
		Format: "${time} ${status} - ${method} ${path} - ${latency}\n",
	}))

	// Security middleware
	app.Server.Use(helmet.New())

	// CORS middleware
	app.Server.Use(cors.New(cors.Config{
		AllowOrigins:     app.Config.Server.AllowedOrigins,
		AllowMethods:     "GET,POST,PUT,DELETE,OPTIONS",
		AllowHeaders:     "Origin,Content-Type,Accept,Authorization,DPoP",
		AllowCredentials: true,
	}))

	// Tracing middleware
	app.Server.Use(tracing.TraceMiddleware("auth-service"))
}

func (app *Application) setupRoutes() {
	// Initialize services and handlers
	userRepo := repository.NewUserRepository(app.DB)
	tokenRepo := repository.NewTokenRepository(app.DB)
	
	// Initialize HSM client (mock for development)
	hsmClient := hsm.NewMockHSMClient(&app.Config.HSM)
	keyManager := hsm.NewKeyManager(hsmClient, &app.Config.HSM)
	
	// Get private key for JWT service
	privateKey, err := keyManager.GetCurrentPrivateKey()
	if err != nil {
		app.Logger.Error("Failed to get private key", "error", err)
		os.Exit(1)
	}
	
	// Initialize JWT service
	jwtService := jwt.NewJWTService(&app.Config.JWT, privateKey)
	
	// Initialize WhatsApp service
	var whatsappClient whatsapp.WhatsAppClient
	if app.Config.Environment == "development" {
		whatsappClient = whatsapp.NewMockWhatsAppClient()
	} else {
		whatsappClient = whatsapp.NewWhatsAppClient(&app.Config.WhatsApp)
	}
	whatsappService := whatsapp.NewWhatsAppService(whatsappClient)
	
	// Initialize auth service
	authService := service.NewAuthService(
		app.Config,
		userRepo,
		tokenRepo,
		jwtService,
		whatsappService,
		app.Redis,
		app.Logger,
	)
	
	// Initialize handlers
	authHandler := handlers.NewAuthHandler(authService, app.Logger)
	jwksHandler := handlers.NewJWKSHandler(jwtService)
	
	// Health check
	app.Server.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status":    "healthy",
			"timestamp": time.Now().UTC(),
			"service":   "auth-service",
			"version":   "1.0.0",
		})
	})

	// Metrics endpoint
	app.Server.Get("/metrics", func(c *fiber.Ctx) error {
		// TODO: Implement Prometheus metrics handler
		return c.SendString("# Metrics endpoint - TODO: Implement Prometheus handler")
	})

	// JWKS and OpenID Connect endpoints
	app.Server.Get("/.well-known/jwks.json", jwksHandler.GetJWKS)
	app.Server.Get("/.well-known/openid-configuration", jwksHandler.GetOpenIDConfiguration)

	// API routes
	api := app.Server.Group("/v1")
	
	// Auth routes
	auth := api.Group("/auth")
	auth.Post("/register", authHandler.Register)
	auth.Post("/login", authHandler.LoginWithPassword)
	auth.Post("/login/whatsapp", authHandler.LoginWithWhatsApp)
	auth.Post("/otp/send", authHandler.SendWhatsAppOTP)
	auth.Post("/refresh", authHandler.RefreshToken)
	auth.Post("/logout", middleware.AuthMiddleware(jwtService), authHandler.Logout)
	auth.Post("/revoke", middleware.AuthMiddleware(jwtService), authHandler.RevokeToken)
	
	// User routes
	user := api.Group("/user", middleware.AuthMiddleware(jwtService))
	user.Get("/profile", authHandler.GetProfile)
	user.Get("/:user_id/permissions", authHandler.GetUserPermissions)
	
	// Tenant routes
	tenants := api.Group("/tenants", middleware.AuthMiddleware(jwtService))
	tenants.Post("/:tenant_id/transfer-president", authHandler.TransferPresident)
	tenants.Get("/:tenant_id/president", authHandler.GetTenantPresident)

	// OAuth routes placeholder
	oauth := app.Server.Group("/oauth")
	oauth.Get("/", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"message": "OAuth 2.1 + OIDC endpoints",
			"status":  "ready",
		})
	})
}