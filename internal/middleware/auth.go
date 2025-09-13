package middleware

import (
	"strings"

	"github.com/smartedify/auth-service/internal/errors"
	"github.com/smartedify/auth-service/internal/jwt"

	"github.com/gofiber/fiber/v2"
)

// AuthMiddleware validates JWT tokens
func AuthMiddleware(jwtService jwt.JWTService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Get token from Authorization header
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(errors.ErrTokenInvalid.WithDetails("missing authorization header"))
		}
		
		// Check Bearer prefix
		if !strings.HasPrefix(authHeader, "Bearer ") {
			return c.Status(fiber.StatusUnauthorized).JSON(errors.ErrTokenInvalid.WithDetails("invalid authorization header format"))
		}
		
		// Extract token
		tokenString := authHeader[7:]
		if tokenString == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(errors.ErrTokenInvalid.WithDetails("missing token"))
		}
		
		// Validate token
		claims, err := jwtService.ValidateToken(tokenString)
		if err != nil {
			if apiErr, ok := errors.IsAPIError(err); ok {
				return c.Status(apiErr.Status).JSON(apiErr)
			}
			return c.Status(fiber.StatusUnauthorized).JSON(errors.ErrTokenInvalid.WithDetails(err.Error()))
		}
		
		// Store claims in context
		c.Locals("user_id", claims.Subject)
		c.Locals("tenant_id", claims.TenantID)
		c.Locals("unit_id", claims.UnitID)
		c.Locals("jti", claims.JTI)
		c.Locals("claims", claims)
		
		return c.Next()
	}
}

// OptionalAuthMiddleware validates JWT tokens but doesn't require them
func OptionalAuthMiddleware(jwtService jwt.JWTService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return c.Next()
		}
		
		if !strings.HasPrefix(authHeader, "Bearer ") {
			return c.Next()
		}
		
		tokenString := authHeader[7:]
		if tokenString == "" {
			return c.Next()
		}
		
		claims, err := jwtService.ValidateToken(tokenString)
		if err != nil {
			// Don't return error for optional auth, just continue without user context
			return c.Next()
		}
		
		// Store claims in context if valid
		c.Locals("user_id", claims.Subject)
		c.Locals("tenant_id", claims.TenantID)
		c.Locals("unit_id", claims.UnitID)
		c.Locals("jti", claims.JTI)
		c.Locals("claims", claims)
		
		return c.Next()
	}
}

// RequireRole middleware checks if user has required role
func RequireRole(requiredRole string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		userID := c.Locals("user_id")
		tenantID := c.Locals("tenant_id")
		unitID := c.Locals("unit_id")
		
		if userID == nil || tenantID == nil || unitID == nil {
			return c.Status(fiber.StatusUnauthorized).JSON(errors.ErrTokenInvalid.WithDetails("missing user context"))
		}
		
		// TODO: Query user permissions from database
		// This would typically involve calling the user repository
		// For now, we'll assume the role check passes
		
		return c.Next()
	}
}

// RequirePresident middleware checks if user is president
func RequirePresident() fiber.Handler {
	return func(c *fiber.Ctx) error {
		userID := c.Locals("user_id")
		tenantID := c.Locals("tenant_id")
		
		if userID == nil || tenantID == nil {
			return c.Status(fiber.StatusUnauthorized).JSON(errors.ErrTokenInvalid.WithDetails("missing user context"))
		}
		
		// TODO: Check if user is president of the tenant
		// This would involve querying the tenant_presidents table
		
		return c.Next()
	}
}

// RequireOwner middleware checks if user is an owner
func RequireOwner() fiber.Handler {
	return RequireRole("owner")
}

// ExtractUserContext helper function to get user context from fiber context
func ExtractUserContext(c *fiber.Ctx) (userID, tenantID, unitID string, ok bool) {
	userIDVal := c.Locals("user_id")
	tenantIDVal := c.Locals("tenant_id")
	unitIDVal := c.Locals("unit_id")
	
	if userIDVal == nil {
		return "", "", "", false
	}
	
	userID, ok = userIDVal.(string)
	if !ok {
		return "", "", "", false
	}
	
	if tenantIDVal != nil {
		tenantID, _ = tenantIDVal.(string)
	}
	
	if unitIDVal != nil {
		unitID, _ = unitIDVal.(string)
	}
	
	return userID, tenantID, unitID, true
}

// ExtractClaims helper function to get JWT claims from fiber context
func ExtractClaims(c *fiber.Ctx) (*jwt.TokenClaims, bool) {
	claimsVal := c.Locals("claims")
	if claimsVal == nil {
		return nil, false
	}
	
	claims, ok := claimsVal.(*jwt.TokenClaims)
	return claims, ok
}