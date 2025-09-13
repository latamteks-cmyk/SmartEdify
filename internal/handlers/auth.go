package handlers

import (
	"log/slog"

	"github.com/smartedify/auth-service/internal/errors"
	"github.com/smartedify/auth-service/internal/models"
	"github.com/smartedify/auth-service/internal/server"
	"github.com/smartedify/auth-service/internal/service"

	"github.com/gofiber/fiber/v2"
)

// AuthHandler handles authentication endpoints
type AuthHandler struct {
	authService service.AuthService
	logger      *slog.Logger
}

func NewAuthHandler(authService service.AuthService, logger *slog.Logger) *AuthHandler {
	return &AuthHandler{
		authService: authService,
		logger:      logger,
	}
}

// Register handles user registration
func (h *AuthHandler) Register(c *fiber.Ctx) error {
	var req models.CreateUserRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(errors.ErrInvalidInput.WithDetails("invalid request body"))
	}

	user, err := h.authService.RegisterUser(c.Context(), &req)
	if err != nil {
		if apiErr, ok := errors.IsAPIError(err); ok {
			return c.Status(apiErr.Status).JSON(apiErr)
		}
		return c.Status(fiber.StatusInternalServerError).JSON(errors.ErrInternalServer.WithDetails(err.Error()))
	}

	// Remove sensitive data from response
	user.PasswordHash = ""
	user.MFASecret = ""

	return server.SuccessResponse(c, user)
}

// SendWhatsAppOTP handles sending OTP via WhatsApp
func (h *AuthHandler) SendWhatsAppOTP(c *fiber.Ctx) error {
	var req struct {
		Phone string `json:"phone" validate:"required"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(errors.ErrInvalidInput.WithDetails("invalid request body"))
	}

	// Validar que el campo phone no esté vacío
	if req.Phone == "" {
		return c.Status(fiber.StatusBadRequest).JSON(errors.ErrInvalidInput.WithDetails("phone is required"))
	}

	if err := h.authService.SendWhatsAppOTP(c.Context(), req.Phone); err != nil {
		if apiErr, ok := errors.IsAPIError(err); ok {
			return c.Status(apiErr.Status).JSON(apiErr)
		}
		return c.Status(fiber.StatusInternalServerError).JSON(errors.ErrInternalServer.WithDetails(err.Error()))
	}

	return server.SuccessResponse(c, fiber.Map{
		"message": "OTP sent successfully",
		"phone":   req.Phone,
	})
}

// LoginWithWhatsApp handles WhatsApp OTP login
func (h *AuthHandler) LoginWithWhatsApp(c *fiber.Ctx) error {
	var req models.WhatsAppLoginRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(errors.ErrInvalidInput.WithDetails("invalid request body"))
	}

	response, err := h.authService.LoginWithWhatsApp(c.Context(), &req)
	if err != nil {
		if apiErr, ok := errors.IsAPIError(err); ok {
			return c.Status(apiErr.Status).JSON(apiErr)
		}
		return c.Status(fiber.StatusInternalServerError).JSON(errors.ErrInternalServer.WithDetails(err.Error()))
	}

	// Set secure cookie for refresh token (optional)
	c.Cookie(&fiber.Cookie{
		Name:     "refresh_token",
		Value:    response.RefreshToken,
		HTTPOnly: true,
		Secure:   true,
		SameSite: "Strict",
		MaxAge:   7 * 24 * 60 * 60, // 7 days
	})

	return server.SuccessResponse(c, response)
}

// LoginWithPassword handles email/password login
func (h *AuthHandler) LoginWithPassword(c *fiber.Ctx) error {
	var req models.LoginRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(errors.ErrInvalidInput.WithDetails("invalid request body"))
	}

	response, err := h.authService.LoginWithPassword(c.Context(), &req)
	if err != nil {
		if apiErr, ok := errors.IsAPIError(err); ok {
			return c.Status(apiErr.Status).JSON(apiErr)
		}
		return c.Status(fiber.StatusInternalServerError).JSON(errors.ErrInternalServer.WithDetails(err.Error()))
	}

	// Set secure cookie for refresh token
	c.Cookie(&fiber.Cookie{
		Name:     "refresh_token",
		Value:    response.RefreshToken,
		HTTPOnly: true,
		Secure:   true,
		SameSite: "Strict",
		MaxAge:   7 * 24 * 60 * 60, // 7 days
	})

	return server.SuccessResponse(c, response)
}

// RefreshToken handles token refresh
func (h *AuthHandler) RefreshToken(c *fiber.Ctx) error {
	var req models.RefreshTokenRequest

	// Try to get refresh token from body first
	if err := c.BodyParser(&req); err != nil {
		// If body parsing fails, try to get from cookie
		refreshToken := c.Cookies("refresh_token")
		if refreshToken == "" {
			return c.Status(fiber.StatusBadRequest).JSON(errors.ErrInvalidInput.WithDetails("refresh token required"))
		}
		req.RefreshToken = refreshToken
	}

	response, err := h.authService.RefreshToken(c.Context(), &req)
	if err != nil {
		if apiErr, ok := errors.IsAPIError(err); ok {
			return c.Status(apiErr.Status).JSON(apiErr)
		}
		return c.Status(fiber.StatusInternalServerError).JSON(errors.ErrInternalServer.WithDetails(err.Error()))
	}

	return server.SuccessResponse(c, response)
}

// Logout handles user logout
func (h *AuthHandler) Logout(c *fiber.Ctx) error {
	// Get JTI from token claims (set by auth middleware)
	jti := c.Locals("jti")
	if jti == nil {
		return c.Status(fiber.StatusBadRequest).JSON(errors.ErrTokenInvalid.WithDetails("invalid token"))
	}

	jtiStr, ok := jti.(string)
	if !ok {
		return c.Status(fiber.StatusBadRequest).JSON(errors.ErrTokenInvalid.WithDetails("invalid token"))
	}

	if err := h.authService.RevokeToken(c.Context(), jtiStr); err != nil {
		h.logger.Warn("Failed to revoke token during logout", "error", err, "jti", jtiStr)
	}

	// Clear refresh token cookie
	c.Cookie(&fiber.Cookie{
		Name:     "refresh_token",
		Value:    "",
		HTTPOnly: true,
		Secure:   true,
		SameSite: "Strict",
		MaxAge:   -1, // Delete cookie
	})

	return server.SuccessResponse(c, fiber.Map{
		"message": "Logged out successfully",
	})
}

// RevokeToken handles token revocation
func (h *AuthHandler) RevokeToken(c *fiber.Ctx) error {
	var req struct {
		JTI string `json:"jti" validate:"required"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(errors.ErrInvalidInput.WithDetails("invalid request body"))
	}

	if err := h.authService.RevokeToken(c.Context(), req.JTI); err != nil {
		if apiErr, ok := errors.IsAPIError(err); ok {
			return c.Status(apiErr.Status).JSON(apiErr)
		}
		return c.Status(fiber.StatusInternalServerError).JSON(errors.ErrInternalServer.WithDetails(err.Error()))
	}

	return server.SuccessResponse(c, fiber.Map{
		"message": "Token revoked successfully",
	})
}

// GetProfile returns the current user's profile
func (h *AuthHandler) GetProfile(c *fiber.Ctx) error {
	userID := c.Locals("user_id")
	if userID == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(errors.ErrTokenInvalid.WithDetails("missing user context"))
	}

	userIDStr, ok := userID.(string)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(errors.ErrTokenInvalid.WithDetails("invalid user context"))
	}

	user, err := h.authService.GetUserByID(c.Context(), userIDStr)
	if err != nil {
		if apiErr, ok := errors.IsAPIError(err); ok {
			return c.Status(apiErr.Status).JSON(apiErr)
		}
		return c.Status(fiber.StatusInternalServerError).JSON(errors.ErrInternalServer.WithDetails(err.Error()))
	}

	// Remove sensitive data
	user.PasswordHash = ""
	user.MFASecret = ""

	return server.SuccessResponse(c, user)
}

// GetUserPermissions returns user permissions for a specific tenant/unit
func (h *AuthHandler) GetUserPermissions(c *fiber.Ctx) error {
	userID := c.Params("user_id")
	tenantID := c.Query("tenant_id")
	unitID := c.Query("unit_id")

	if userID == "" || tenantID == "" || unitID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(errors.ErrMissingRequired.WithDetails("user_id, tenant_id, and unit_id are required"))
	}

	permissions, err := h.authService.GetUserPermissions(c.Context(), userID, tenantID, unitID)
	if err != nil {
		if apiErr, ok := errors.IsAPIError(err); ok {
			return c.Status(apiErr.Status).JSON(apiErr)
		}
		return c.Status(fiber.StatusInternalServerError).JSON(errors.ErrInternalServer.WithDetails(err.Error()))
	}

	return server.SuccessResponse(c, permissions)
}

// TransferPresident handles president transfer
func (h *AuthHandler) TransferPresident(c *fiber.Ctx) error {
	tenantID := c.Params("tenant_id")
	if tenantID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(errors.ErrMissingRequired.WithDetails("tenant_id is required"))
	}

	var req models.TransferPresidentRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(errors.ErrInvalidInput.WithDetails("invalid request body"))
	}

	// Get current user from token
	fromUserID := c.Locals("user_id")
	if fromUserID == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(errors.ErrTokenInvalid.WithDetails("missing user context"))
	}

	fromUserIDStr, ok := fromUserID.(string)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(errors.ErrTokenInvalid.WithDetails("invalid user context"))
	}

	if err := h.authService.TransferPresident(c.Context(), tenantID, req.ToUserID, fromUserIDStr); err != nil {
		if apiErr, ok := errors.IsAPIError(err); ok {
			return c.Status(apiErr.Status).JSON(apiErr)
		}
		return c.Status(fiber.StatusInternalServerError).JSON(errors.ErrInternalServer.WithDetails(err.Error()))
	}

	return server.SuccessResponse(c, fiber.Map{
		"message":      "President transfer initiated successfully",
		"tenant_id":    tenantID,
		"to_user_id":   req.ToUserID,
		"from_user_id": fromUserIDStr,
	})
}

// GetTenantPresident returns the current president of a tenant
func (h *AuthHandler) GetTenantPresident(c *fiber.Ctx) error {
	tenantID := c.Params("tenant_id")
	if tenantID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(errors.ErrMissingRequired.WithDetails("tenant_id is required"))
	}

	president, err := h.authService.GetTenantPresident(c.Context(), tenantID)
	if err != nil {
		if apiErr, ok := errors.IsAPIError(err); ok {
			return c.Status(apiErr.Status).JSON(apiErr)
		}
		return c.Status(fiber.StatusInternalServerError).JSON(errors.ErrInternalServer.WithDetails(err.Error()))
	}

	return server.SuccessResponse(c, president)
}
