package middleware

import (
	"log/slog"
	"strings"
	"time"

	"github.com/smartedify/auth-service/internal/errors"

	"github.com/gofiber/fiber/v2"
)

// RequestID middleware adds a unique request ID to each request
func RequestID() fiber.Handler {
	return func(c *fiber.Ctx) error {
		requestID := c.Get("X-Request-ID")
		if requestID == "" {
			requestID = generateRequestID()
		}
		
		c.Set("X-Request-ID", requestID)
		c.Locals("requestID", requestID)
		
		return c.Next()
	}
}

// SecurityHeaders middleware adds security headers
func SecurityHeaders() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Security headers
		c.Set("X-Content-Type-Options", "nosniff")
		c.Set("X-Frame-Options", "DENY")
		c.Set("X-XSS-Protection", "1; mode=block")
		c.Set("Referrer-Policy", "strict-origin-when-cross-origin")
		c.Set("Permissions-Policy", "geolocation=(), microphone=(), camera=()")
		
		// Remove server header
		c.Set("Server", "")
		
		return c.Next()
	}
}

// CORS middleware with proper configuration
func CORS(allowedOrigins []string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		origin := c.Get("Origin")
		
		// Check if origin is allowed
		allowed := false
		for _, allowedOrigin := range allowedOrigins {
			if allowedOrigin == "*" || allowedOrigin == origin {
				allowed = true
				break
			}
		}
		
		if allowed {
			c.Set("Access-Control-Allow-Origin", origin)
		}
		
		c.Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Set("Access-Control-Allow-Headers", "Origin, Content-Type, Accept, Authorization, DPoP, X-Request-ID")
		c.Set("Access-Control-Allow-Credentials", "true")
		c.Set("Access-Control-Max-Age", "86400")
		
		// Handle preflight requests
		if c.Method() == "OPTIONS" {
			return c.SendStatus(fiber.StatusNoContent)
		}
		
		return c.Next()
	}
}

// RateLimit middleware for basic rate limiting
func RateLimit(maxRequests int, window time.Duration) fiber.Handler {
	// This is a simple in-memory rate limiter
	// In production, use Redis-based rate limiting
	return func(c *fiber.Ctx) error {
		// TODO: Implement proper rate limiting with Redis
		// For now, just pass through
		return c.Next()
	}
}

// ValidateContentType middleware validates request content type
func ValidateContentType(allowedTypes ...string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		if c.Method() == "GET" || c.Method() == "DELETE" {
			return c.Next()
		}
		
		contentType := c.Get("Content-Type")
		if contentType == "" {
			return c.Status(fiber.StatusBadRequest).JSON(errors.ErrInvalidFormat.WithDetails("Content-Type header is required"))
		}
		
		// Extract media type (ignore charset and other parameters)
		mediaType := strings.Split(contentType, ";")[0]
		mediaType = strings.TrimSpace(mediaType)
		
		for _, allowedType := range allowedTypes {
			if mediaType == allowedType {
				return c.Next()
			}
		}
		
		return c.Status(fiber.StatusUnsupportedMediaType).JSON(errors.ErrInvalidFormat.WithDetails("Unsupported content type"))
	}
}

// RequestLogger middleware logs requests
func RequestLogger() fiber.Handler {
	return func(c *fiber.Ctx) error {
		start := time.Now()
		
		err := c.Next()
		
		duration := time.Since(start)
		
		slog.Info("Request completed",
			"method", c.Method(),
			"path", c.Path(),
			"status", c.Response().StatusCode(),
			"duration", duration.String(),
			"ip", c.IP(),
			"user_agent", c.Get("User-Agent"),
			"request_id", c.Locals("requestID"),
		)
		
		return err
	}
}

// ErrorHandler middleware handles panics and errors
func ErrorHandler() fiber.Handler {
	return func(c *fiber.Ctx) error {
		defer func() {
			if r := recover(); r != nil {
				slog.Error("Panic recovered",
					"panic", r,
					"path", c.Path(),
					"method", c.Method(),
					"ip", c.IP(),
				)
				
				c.Status(fiber.StatusInternalServerError).JSON(errors.ErrInternalServer)
			}
		}()
		
		return c.Next()
	}
}

// generateRequestID generates a unique request ID
func generateRequestID() string {
	// Simple implementation - in production use UUID or similar
	return time.Now().Format("20060102150405") + "-" + randomString(8)
}

func randomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyz0123456789"
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[time.Now().UnixNano()%int64(len(charset))]
	}
	return string(b)
}