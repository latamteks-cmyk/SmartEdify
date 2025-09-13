package handlers

import (
	"github.com/smartedify/auth-service/internal/jwt"
	"github.com/smartedify/auth-service/internal/server"

	"github.com/gofiber/fiber/v2"
)

// JWKSHandler handles JWKS endpoint
type JWKSHandler struct {
	jwtService jwt.JWTService
}

func NewJWKSHandler(jwtService jwt.JWTService) *JWKSHandler {
	return &JWKSHandler{
		jwtService: jwtService,
	}
}

// GetJWKS returns the JSON Web Key Set
func (h *JWKSHandler) GetJWKS(c *fiber.Ctx) error {
	// Set cache headers for JWKS
	c.Set("Cache-Control", "public, max-age=3600") // Cache for 1 hour
	c.Set("Content-Type", "application/json")
	
	// Generate JWKS from current public key
	jwks, err := jwt.GenerateJWKSFromService(h.jwtService)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to generate JWKS",
		})
	}
	
	return server.SuccessResponse(c, jwks)
}

// GetOpenIDConfiguration returns OpenID Connect configuration
func (h *JWKSHandler) GetOpenIDConfiguration(c *fiber.Ctx) error {
	// Set cache headers
	c.Set("Cache-Control", "public, max-age=3600")
	c.Set("Content-Type", "application/json")
	
	baseURL := c.BaseURL()
	
	config := fiber.Map{
		"issuer":                 baseURL,
		"authorization_endpoint": baseURL + "/oauth/authorize",
		"token_endpoint":         baseURL + "/oauth/token",
		"userinfo_endpoint":      baseURL + "/oauth/userinfo",
		"jwks_uri":              baseURL + "/.well-known/jwks.json",
		"introspection_endpoint": baseURL + "/oauth/introspect",
		"revocation_endpoint":    baseURL + "/oauth/revoke",
		"response_types_supported": []string{
			"code",
			"token",
			"id_token",
			"code token",
			"code id_token",
			"token id_token",
			"code token id_token",
		},
		"subject_types_supported": []string{"public"},
		"id_token_signing_alg_values_supported": []string{"RS256"},
		"scopes_supported": []string{
			"openid",
			"profile",
			"email",
			"phone",
			"read:units",
			"write:units",
			"read:assemblies",
			"write:assemblies",
			"read:payments",
			"write:payments",
		},
		"token_endpoint_auth_methods_supported": []string{
			"client_secret_basic",
			"client_secret_post",
			"private_key_jwt",
		},
		"claims_supported": []string{
			"sub",
			"iss",
			"aud",
			"exp",
			"iat",
			"jti",
			"tenant_id",
			"unit_id",
			"name",
			"email",
			"phone",
		},
		"grant_types_supported": []string{
			"authorization_code",
			"refresh_token",
			"client_credentials",
		},
		"code_challenge_methods_supported": []string{"S256"},
		"dpop_signing_alg_values_supported": []string{"RS256", "ES256"},
	}
	
	return server.SuccessResponse(c, config)
}