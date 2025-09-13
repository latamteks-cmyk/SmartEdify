package service

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/smartedify/auth-service/internal/config"
	"github.com/smartedify/auth-service/internal/errors"
	"github.com/smartedify/auth-service/internal/jwt"
	"github.com/smartedify/auth-service/internal/models"
	"github.com/smartedify/auth-service/internal/repository"
	"github.com/smartedify/auth-service/internal/utils"
	"github.com/smartedify/auth-service/internal/whatsapp"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/valyala/fasthttp"
)

// AuthService interface defines authentication operations
type AuthService interface {
	// User registration and management
	RegisterUser(ctx *fasthttp.RequestCtx, req *models.CreateUserRequest) (*models.User, error)
	GetUserByID(ctx *fasthttp.RequestCtx, userID string) (*models.User, error)
	UpdateUserStatus(ctx *fasthttp.RequestCtx, userID, status string) error
	SendWhatsAppOTP(ctx *fasthttp.RequestCtx, phone string) error
	RevokeToken(ctx *fasthttp.RequestCtx, jti string) error
	RevokeAllUserTokens(ctx *fasthttp.RequestCtx, userID string) error
	// User permissions
	GetUserPermissions(ctx *fasthttp.RequestCtx, userID, tenantID, unitID string) (*models.UserPermissions, error)
	// President management
	TransferPresident(ctx *fasthttp.RequestCtx, tenantID, toUserID, fromUserID string) error
	// Métodos de login y refresh para la interfaz AuthService
	LoginWithWhatsApp(ctx *fasthttp.RequestCtx, req *models.WhatsAppLoginRequest) (*models.LoginResponse, error)
	LoginWithPassword(ctx *fasthttp.RequestCtx, req *models.LoginRequest) (*models.LoginResponse, error)
	RefreshToken(ctx *fasthttp.RequestCtx, req *models.RefreshTokenRequest) (*models.LoginResponse, error)
	// Gestión de presidente de tenant
	GetTenantPresident(ctx *fasthttp.RequestCtx, tenantID string) (*models.TenantPresident, error)
}

// NewAuthService creates a new authentication service
// authService implementa AuthService
type authService struct {
	config      *config.Config
	userRepo    repository.UserRepository
	tokenRepo   repository.TokenRepository
	jwtService  jwt.JWTService
	whatsappSvc *whatsapp.WhatsAppService
	redis       *redis.Client
	logger      *slog.Logger
}

func NewAuthService(
	cfg *config.Config,
	userRepo repository.UserRepository,
	tokenRepo repository.TokenRepository,
	jwtService jwt.JWTService,
	whatsappSvc *whatsapp.WhatsAppService,
	redis *redis.Client,
	logger *slog.Logger,
) AuthService {
	return &authService{
		config:      cfg,
		userRepo:    userRepo,
		tokenRepo:   tokenRepo,
		jwtService:  jwtService,
		whatsappSvc: whatsappSvc,
		redis:       redis,
		logger:      logger,
	}
}

func (s *authService) RegisterUser(ctx *fasthttp.RequestCtx, req *models.CreateUserRequest) (*models.User, error) {
	// Validar email
	if err := utils.ValidateEmail(req.Email); err != nil {
		return nil, errors.ErrInvalidInput.WithDetails("INVALID_EMAIL_FORMAT")
	}
	// Validar nombre
	if err := utils.ValidateName(req.Name, "Nombre"); err != nil {
		return nil, errors.ErrInvalidInput.WithDetails("MISSING_REQUIRED_FIELDS")
	}
	// Validar tenant
	if err := utils.ValidateUUID(req.TenantID); err != nil {
		return nil, errors.ErrInvalidInput.WithDetails("MISSING_REQUIRED_FIELDS")
	}
	// Validar aceptación de términos
	if !req.TermsAccepted {
		return nil, errors.ErrInvalidInput.WithDetails("TERMS_NOT_ACCEPTED")
	}
	// Validar existencia de email en tenant
	if exists, err := s.userRepo.EmailExists(ctx, req.Email); err != nil {
		return nil, err
	} else if exists {
		return nil, errors.ErrEmailAlreadyExists.WithDetails("EMAIL_ALREADY_EXISTS")
	}
	// Validar contraseña
	if req.Password == "" {
		return nil, errors.ErrInvalidInput.WithDetails("MISSING_REQUIRED_FIELDS")
	}
	if err := ValidatePassword(req.Password); err != nil {
		return nil, errors.ErrInvalidInput.WithDetails("WEAK_PASSWORD: " + err.Error())
	}
	// Crear usuario
	user := &models.User{
		ID:            uuid.New(),
		Email:         req.Email,
		Phone:         req.Phone,
		Name:          req.Name,
		Status:        models.UserStatusInactive,
		EmailVerified: false,
		PhoneVerified: false,
		MFAEnabled:    false,
	}
	// Hash de contraseña
	hashedPassword, err := HashPassword(req.Password)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}
	user.PasswordHash = hashedPassword
	// Guardar usuario
	if err := s.userRepo.CreateUser(ctx, user); err != nil {
		return nil, err
	}

	// Audit logging
	// Registrar evento de auditoría por creación de usuario
	// Se asume que el paquete monitoring está importado correctamente
	// El campo UserID puede ser el ID del usuario creado
	// El campo Action puede ser "register_user"
	// El campo Resource puede ser el email
	// El campo Success es true
	// El campo Details puede incluir el tenant
	// Si el usuario se crea correctamente, se registra el evento
	// Si hay error, se puede registrar un evento de error (no incluido aquí por simplicidad)
	// monitoring.LogAuditEvent(monitoring.NewAuditEvent(user.ID.String(), "register_user", user.Email, true, "tenant_id="+req.TenantID))

	return user, nil
}

func (s *authService) GetUserByID(ctx *fasthttp.RequestCtx, userID string) (*models.User, error) {
	if err := utils.ValidateUUID(userID); err != nil {
		// Audit logging: intento fallido de consulta por UUID inválido
		// monitoring.LogAuditEvent(monitoring.NewAuditEvent(userID, "get_user_by_id", "", false, "invalid UUID"))
		return nil, err
	}
	user, err := s.userRepo.GetUserByID(ctx, userID)
	if err != nil {
		// Audit logging: intento fallido de consulta por error en repositorio
		// monitoring.LogAuditEvent(monitoring.NewAuditEvent(userID, "get_user_by_id", "", false, "repo error: "+err.Error()))
		return nil, err
	}
	// Audit logging: consulta exitosa
	// monitoring.LogAuditEvent(monitoring.NewAuditEvent(userID, "get_user_by_id", "", true, "user retrieved"))
	return user, nil
}

func (s *authService) UpdateUserStatus(ctx *fasthttp.RequestCtx, userID, status string) error {
	if err := utils.ValidateUUID(userID); err != nil {
		// Audit logging: intento fallido de cambio de estado por UUID inválido
		// monitoring.LogAuditEvent(monitoring.NewAuditEvent(userID, "update_user_status", "", false, "invalid UUID"))
		return err
	}
	validStatuses := map[string]bool{
		"pending_verification": true,
		"active":               true,
		"suspended":            true,
		"deleted":              true,
	}
	if !validStatuses[status] {
		// Audit logging: intento fallido de cambio de estado por status inválido
		// monitoring.LogAuditEvent(monitoring.NewAuditEvent(userID, "update_user_status", "", false, "invalid status: "+status))
		return errors.ErrInvalidInput.WithDetails("invalid status")
	}
	if err := s.userRepo.UpdateUserStatus(ctx, userID, status); err != nil {
		// Audit logging: intento fallido de cambio de estado por error en repositorio
		// monitoring.LogAuditEvent(monitoring.NewAuditEvent(userID, "update_user_status", "", false, "repo error: "+err.Error()))
		return err
	}
	// Audit logging: cambio de estado exitoso
	// monitoring.LogAuditEvent(monitoring.NewAuditEvent(userID, "update_user_status", "", true, "status updated to "+status))
	return nil
}

func (s *authService) SendWhatsAppOTP(ctx *fasthttp.RequestCtx, phone string) error {
	// Check rate limiting
	rateLimitKey := fmt.Sprintf("otp_rate_limit:%s", phone)
	count, err := s.redis.Get(ctx, rateLimitKey).Int()
	if err != nil && err != redis.Nil {
		return fmt.Errorf("failed to check rate limit: %w", err)
	}

	if count >= 3 { // Max 3 OTP requests per hour
		// Audit logging: intento fallido por rate limit
		// monitoring.LogAuditEvent(monitoring.NewAuditEvent("", "send_otp", phone, false, "rate limit exceeded"))
		return errors.ErrTooManyAttempts.WithDetails("too many OTP requests")
	}

	// Generate OTP
	otp, _ := utils.GenerateOTP(6)
	// Hash OTP for storage
	otpHash, err := utils.HashOTP(otp)
	if err != nil {
		return fmt.Errorf("failed to hash OTP: %w", err)
	}

	// Store OTP in database
	otpCode := &models.OTPCode{
		ID:          uuid.New().String(),
		Phone:       phone,
		Code:        otp,
		CodeHash:    otpHash,
		Purpose:     "login",
		ExpiresAt:   time.Now().Add(5 * time.Minute),
		MaxAttempts: 3,
	}

	if err := s.tokenRepo.CreateOTPCode(ctx, otpCode); err != nil {
		// Audit logging: intento fallido por error al guardar OTP
		// monitoring.LogAuditEvent(monitoring.NewAuditEvent("", "send_otp", phone, false, "repo error: "+err.Error()))
		return err
	}

	// Send OTP via WhatsApp
	if err := s.whatsappSvc.SendLoginOTP(ctx, phone, otp); err != nil {
		s.logger.Error("Failed to send WhatsApp OTP",
			"phone", phone,
			"error", err,
		)
		// Audit logging: intento fallido por error al enviar OTP
		// monitoring.LogAuditEvent(monitoring.NewAuditEvent("", "send_otp", phone, false, "send error: "+err.Error()))
		return err
	}

	// Update rate limit
	if err := s.redis.Set(ctx, rateLimitKey, count+1, time.Hour).Err(); err != nil {
		s.logger.Warn("Failed to update rate limit", "error", err)
	}

	s.logger.Info("WhatsApp OTP sent successfully", "phone", phone)
	// Audit logging: envío exitoso de OTP
	// monitoring.LogAuditEvent(monitoring.NewAuditEvent("", "send_otp", phone, true, "OTP sent"))
	return nil
}

func (s *authService) LoginWithWhatsApp(ctx *fasthttp.RequestCtx, req *models.WhatsAppLoginRequest) (*models.LoginResponse, error) {
	if req.Phone == "" {
		return nil, errors.ErrMissingRequired.WithDetails("phone required")
	}
	if req.OTP == "" {
		return nil, errors.ErrMissingRequired.WithDetails("OTP required")
	}

	// Get OTP from database
	otpCode, err := s.tokenRepo.GetOTPCode(ctx, req.Phone, "login")
	if err != nil {
		return nil, errors.ErrOTPInvalid
	}

	// Check if OTP is expired
	if time.Now().After(otpCode.ExpiresAt) {
		return nil, errors.ErrOTPExpired
	}

	// Check if OTP is already used
	if otpCode.Used {
		// Audit logging: intento fallido de login por OTP ya usado
		// monitoring.LogAuditEvent(monitoring.NewAuditEvent("", "login_whatsapp", req.Phone, false, "OTP already used"))
		return nil, errors.ErrOTPAlreadyUsed
	}

	// Check max attempts
	if otpCode.Attempts >= otpCode.MaxAttempts {
		return nil, errors.ErrOTPMaxAttempts
	}

	// Verify OTP
	if !utils.VerifyOTP(req.OTP, otpCode.CodeHash) {
		// Audit logging: intento fallido de login por OTP inválido
		// monitoring.LogAuditEvent(monitoring.NewAuditEvent("", "login_whatsapp", req.Phone, false, "OTP invalid"))
		return nil, errors.ErrOTPInvalid
	}

	// Mark OTP as used
	if err := s.tokenRepo.UseOTPCode(ctx, otpCode.ID); err != nil {
		return nil, err
	}

	// Obtener usuario y generar tokens
	user, err := s.userRepo.GetUserByPhone(ctx, req.Phone)
	if err != nil {
		return nil, err
	}
	accessToken, _, err := s.jwtService.GenerateAccessToken(user.ID.String(), user.TenantID, user.UnitID)
	if err != nil {
		return nil, err
	}
	refreshToken, _, err := s.jwtService.GenerateRefreshToken()
	if err != nil {
		return nil, err
	}
	// Audit logging: login exitoso
	// monitoring.LogAuditEvent(monitoring.NewAuditEvent(user.ID.String(), "login_whatsapp", req.Phone, true, "login success"))
	resp := &models.LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		UserID:       user.ID.String(),
		TenantID:     user.TenantID,
		UnitID:       user.UnitID,
	}
	return resp, nil
}

func (s *authService) LoginWithPassword(ctx *fasthttp.RequestCtx, req *models.LoginRequest) (*models.LoginResponse, error) {
	// Rate limiting y bloqueo por intentos fallidos
	key := fmt.Sprintf("login_attempts:%s", req.Email)
	blockKey := fmt.Sprintf("login_blocked:%s", req.Email)
	// Verificar si el usuario está bloqueado
	blocked, redisErr := s.redis.Get(ctx, blockKey).Result()
	if redisErr == nil && blocked == "1" {
		return nil, errors.ErrInvalidCredentials.WithDetails("ACCOUNT_TEMPORARILY_BLOCKED")
	}
	// Incrementar intentos
	attempts, _ := s.redis.Incr(ctx, key).Result()
	if attempts == 1 {
		s.redis.Expire(ctx, key, 15*time.Minute)
	}
	if attempts > 5 {
		s.redis.Set(ctx, blockKey, "1", 30*time.Minute)
		return nil, errors.ErrInvalidCredentials.WithDetails("ACCOUNT_TEMPORARILY_BLOCKED")
	}
	// Validate input
	var user *models.User
	var err error

	if req.Email == "" {
		return nil, errors.ErrMissingRequired.WithDetails("email required")
	}
	if err := utils.ValidateEmail(req.Email); err != nil {
		return nil, err
	}
	user, err = s.userRepo.GetUserByEmail(ctx, req.Email)
	if err != nil {
		return nil, errors.ErrInvalidCredentials
	}

	// Check password
	if user.PasswordHash == "" || req.Password == "" {
		return nil, errors.ErrInvalidCredentials
	}
	if !utils.VerifyPassword(req.Password, user.PasswordHash) {
		// Audit logging: intento fallido de login por contraseña incorrecta
		// monitoring.LogAuditEvent(monitoring.NewAuditEvent(user.ID.String(), "login_password", req.Email, false, "invalid password"))
		return nil, errors.ErrInvalidCredentials
	}

	// Check user status
	if user.Status != "active" {
		// Audit logging: intento fallido de login por usuario inactivo
		// monitoring.LogAuditEvent(monitoring.NewAuditEvent(user.ID.String(), "login_password", req.Email, false, "user not active"))
		return nil, errors.ErrUserNotActive
	}

	// Update last login
	tenantID := ""
	unitID := ""
	roles, err := s.userRepo.GetUserRoles(ctx, user.ID.String())
	if err != nil {
		s.logger.Warn("Failed to get user roles", "error", err)
	} else if len(roles) > 0 {
		tenantID = roles[0].TenantID
		unitID = roles[0].UnitID
	}

	if err := s.userRepo.UpdateLastLogin(ctx, user.ID.String(), tenantID, unitID); err != nil {
		s.logger.Warn("Failed to update last login", "error", err)
	}

	// Generate tokens
	accessToken, _, err := s.jwtService.GenerateAccessToken(user.ID.String(), tenantID, unitID)
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	refreshTokenString, jti, err := s.jwtService.GenerateRefreshToken()
	if err != nil {
		return nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	// Store refresh token
	refreshToken := &models.RefreshToken{
		ID:        uuid.New().String(),
		JTI:       jti,
		UserID:    user.ID.String(),
		TenantID:  &tenantID,
		UnitID:    &unitID,
		TokenHash: refreshTokenString,
		ExpiresAt: time.Now().Add(time.Duration(s.config.JWT.RefreshTokenTTL) * time.Second),
		DeviceInfo: map[string]interface{}{
			"login_method": "password",
		},
	}

	if err := s.tokenRepo.CreateRefreshToken(ctx, refreshToken); err != nil {
		s.logger.Warn("Failed to store refresh token", "error", err)
	}

	s.logger.Info("Password login successful",
		"user_id", user.ID.String(),
		"email", user.Email,
	)

	// Audit logging: login exitoso
	// monitoring.LogAuditEvent(monitoring.NewAuditEvent(user.ID.String(), "login_password", req.Email, true, "login success"))

	return &models.LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshTokenString,
		ExpiresIn:    s.config.JWT.AccessTokenTTL,
		TokenType:    "Bearer",
		UserID:       user.ID.String(),
		TenantID:     tenantID,
		UnitID:       unitID,
	}, nil
}

func (s *authService) RefreshToken(ctx *fasthttp.RequestCtx, req *models.RefreshTokenRequest) (*models.LoginResponse, error) {
	if req.RefreshToken == "" {
		// Audit logging: intento fallido de refresh por token vacío
		// monitoring.LogAuditEvent(monitoring.NewAuditEvent("", "refresh_token", "", false, "empty refresh token"))
		return nil, errors.ErrInvalidInput.WithDetails("refresh token required")
	}
	_, err := s.jwtService.ValidateToken(req.RefreshToken)
	if err != nil {
		// Audit logging: intento fallido de refresh por token inválido
		// monitoring.LogAuditEvent(monitoring.NewAuditEvent("", "refresh_token", "", false, "invalid refresh token: "+err.Error()))
		return nil, errors.ErrInvalidInput.WithDetails("invalid refresh token")
	}
	// Audit logging: refresh exitoso
	// monitoring.LogAuditEvent(monitoring.NewAuditEvent(claims.Subject, "refresh_token", "", true, "refresh success"))

	// Get refresh token from database
	refreshToken, err := s.tokenRepo.GetRefreshToken(ctx, req.RefreshToken)
	if err != nil {
		return nil, err
	}

	// Check if token is revoked
	if refreshToken.Revoked {
		return nil, errors.ErrTokenRevoked
	}

	// Check if token is expired
	if time.Now().After(refreshToken.ExpiresAt) {
		return nil, errors.ErrTokenExpired
	}

	// Get user
	user, err := s.userRepo.GetUserByID(ctx, refreshToken.UserID)
	if err != nil {
		return nil, err
	}

	// Check user status
	if user.Status != "active" {
		return nil, errors.ErrUserNotActive
	}

	// Generate new access token
	tenantID := ""
	unitID := ""
	if refreshToken.TenantID != nil {
		tenantID = *refreshToken.TenantID
	}
	if refreshToken.UnitID != nil {
		unitID = *refreshToken.UnitID
	}

	accessToken, _, err := s.jwtService.GenerateAccessToken(user.ID.String(), tenantID, unitID)
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	return &models.LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: req.RefreshToken, // Return same refresh token
		ExpiresIn:    s.config.JWT.AccessTokenTTL,
		TokenType:    "Bearer",
		UserID:       user.ID.String(),
		TenantID:     tenantID,
		UnitID:       unitID,
	}, nil
}

func (s *authService) RevokeToken(ctx *fasthttp.RequestCtx, jti string) error {
	if jti == "" {
		// Audit logging: intento fallido de revocación por JTI vacío
		// monitoring.LogAuditEvent(monitoring.NewAuditEvent("", "revoke_token", jti, false, "empty JTI"))
		return errors.ErrInvalidInput.WithDetails("JTI required")
	}
	if err := s.tokenRepo.RevokeRefreshToken(ctx, jti, "manual revocation"); err != nil {
		// Audit logging: intento fallido de revocación por error en repositorio
		// monitoring.LogAuditEvent(monitoring.NewAuditEvent("", "revoke_token", jti, false, "repo error: "+err.Error()))
		return err
	}
	// Audit logging: revocación exitosa
	// monitoring.LogAuditEvent(monitoring.NewAuditEvent("", "revoke_token", jti, true, "token revoked"))
	return nil
}

func (s *authService) RevokeAllUserTokens(ctx *fasthttp.RequestCtx, userID string) error {
	if userID == "" {
		// Audit logging: intento fallido de revocación masiva por userID vacío
		// monitoring.LogAuditEvent(monitoring.NewAuditEvent(userID, "revoke_all_tokens", "", false, "empty userID"))
		return errors.ErrInvalidInput.WithDetails("userID required")
	}
	if err := s.tokenRepo.RevokeAllUserTokens(ctx, userID, "manual revocation"); err != nil {
		// Audit logging: intento fallido de revocación masiva por error en repositorio
		// monitoring.LogAuditEvent(monitoring.NewAuditEvent(userID, "revoke_all_tokens", "", false, "repo error: "+err.Error()))
		return err
	}
	// Audit logging: revocación masiva exitosa
	// monitoring.LogAuditEvent(monitoring.NewAuditEvent(userID, "revoke_all_tokens", "", true, "all tokens revoked"))
	return nil
}

func (s *authService) GetUserPermissions(ctx *fasthttp.RequestCtx, userID, tenantID, unitID string) (*models.UserPermissions, error) {
	if err := utils.ValidateUUID(userID); err != nil {
		// Audit logging: intento fallido de consulta de permisos por UUID inválido
		// monitoring.LogAuditEvent(monitoring.NewAuditEvent(userID, "get_user_permissions", tenantID+":"+unitID, false, "invalid UUID"))
		return nil, err
	}
	perms, err := s.userRepo.GetUserPermissions(ctx, userID, tenantID, unitID)
	if err != nil {
		// Audit logging: intento fallido de consulta de permisos por error en repositorio
		// monitoring.LogAuditEvent(monitoring.NewAuditEvent(userID, "get_user_permissions", tenantID+":"+unitID, false, "repo error: "+err.Error()))
		return nil, err
	}
	// Audit logging: consulta exitosa de permisos
	// monitoring.LogAuditEvent(monitoring.NewAuditEvent(userID, "get_user_permissions", tenantID+":"+unitID, true, "permissions retrieved"))
	return perms, nil
}

func (s *authService) TransferPresident(ctx *fasthttp.RequestCtx, tenantID, toUserID, fromUserID string) error {
	if err := utils.ValidateUUID(tenantID); err != nil {
		// Audit logging: intento fallido de transferencia por tenantID inválido
		// monitoring.LogAuditEvent(monitoring.NewAuditEvent(fromUserID, "transfer_president", tenantID, false, "invalid tenantID"))
		return err
	}
	if err := utils.ValidateUUID(toUserID); err != nil {
		// Audit logging: intento fallido de transferencia por toUserID inválido
		// monitoring.LogAuditEvent(monitoring.NewAuditEvent(fromUserID, "transfer_president", tenantID, false, "invalid toUserID"))
		return err
	}
	if err := utils.ValidateUUID(fromUserID); err != nil {
		// Audit logging: intento fallido de transferencia por fromUserID inválido
		// monitoring.LogAuditEvent(monitoring.NewAuditEvent(fromUserID, "transfer_president", tenantID, false, "invalid fromUserID"))
		return err
	}
	// Simulación para pruebas: no implementado en el repo
	return errors.ErrInvalidInput.WithDetails("not implemented")
}

func (s *authService) GetTenantPresident(ctx *fasthttp.RequestCtx, tenantID string) (*models.TenantPresident, error) {
	if err := utils.ValidateUUID(tenantID); err != nil {
		// Audit logging: intento fallido de consulta de presidente por tenantID inválido
		// monitoring.LogAuditEvent(monitoring.NewAuditEvent("", "get_tenant_president", tenantID, false, "invalid tenantID"))
		return nil, err
	}
	president, err := s.userRepo.GetTenantPresident(ctx, tenantID)
	if err != nil {
		// Audit logging: intento fallido de consulta de presidente por error en repositorio
		// monitoring.LogAuditEvent(monitoring.NewAuditEvent("", "get_tenant_president", tenantID, false, "repo error: "+err.Error()))
		return nil, err
	}
	// Audit logging: consulta exitosa de presidente
	// monitoring.LogAuditEvent(monitoring.NewAuditEvent("", "get_tenant_president", tenantID, true, "president retrieved"))
	return president, nil
}
