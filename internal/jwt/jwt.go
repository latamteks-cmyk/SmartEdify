package jwt

import (
	"crypto/rsa"
	"encoding/base64"
	"fmt"
	"math/big"
	"time"

	"github.com/smartedify/auth-service/internal/config"
	"github.com/smartedify/auth-service/internal/errors"
	"github.com/smartedify/auth-service/internal/models"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// TokenClaims represents the JWT claims structure
type TokenClaims struct {
	Subject   string `json:"sub"`
	TenantID  string `json:"tenant_id,omitempty"`
	UnitID    string `json:"unit_id,omitempty"`
	JTI       string `json:"jti"`
	IssuedAt  int64  `json:"iat"`
	ExpiresAt int64  `json:"exp"`
	Issuer    string `json:"iss"`
	Audience  string `json:"aud"`
	jwt.RegisteredClaims
}

// JWTService handles JWT token operations
type JWTService interface {
	GenerateAccessToken(userID, tenantID, unitID string) (string, *TokenClaims, error)
	GenerateRefreshToken() (string, string, error) // token, jti
	ValidateToken(tokenString string) (*TokenClaims, error)
	GetPublicKey() *rsa.PublicKey
	GetKeyID() string
	RotateKeys() error
}

type jwtService struct {
	config     *config.JWTConfig
	privateKey *rsa.PrivateKey
	publicKey  *rsa.PublicKey
	keyID      string
}

// NewJWTService creates a new JWT service
func NewJWTService(cfg *config.JWTConfig, privateKey *rsa.PrivateKey) JWTService {
	return &jwtService{
		config:     cfg,
		privateKey: privateKey,
		publicKey:  &privateKey.PublicKey,
		keyID:      cfg.KeyVersion,
	}
}

func (s *jwtService) GenerateAccessToken(userID, tenantID, unitID string) (string, *TokenClaims, error) {
	now := time.Now()
	jti := uuid.New().String()
	
	claims := &TokenClaims{
		Subject:   userID,
		TenantID:  tenantID,
		UnitID:    unitID,
		JTI:       jti,
		IssuedAt:  now.Unix(),
		ExpiresAt: now.Add(time.Duration(s.config.AccessTokenTTL) * time.Second).Unix(),
		Issuer:    s.config.Issuer,
		Audience:  s.config.Audience,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   userID,
			ID:        jti,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(time.Duration(s.config.AccessTokenTTL) * time.Second)),
			Issuer:    s.config.Issuer,
			Audience:  []string{s.config.Audience},
		},
	}
	
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	token.Header["kid"] = s.keyID
	
	tokenString, err := token.SignedString(s.privateKey)
	if err != nil {
		return "", nil, fmt.Errorf("failed to sign token: %w", err)
	}
	
	return tokenString, claims, nil
}

func (s *jwtService) GenerateRefreshToken() (string, string, error) {
	jti := uuid.New().String()
	now := time.Now()
	
	claims := &jwt.RegisteredClaims{
		ID:        jti,
		IssuedAt:  jwt.NewNumericDate(now),
		ExpiresAt: jwt.NewNumericDate(now.Add(time.Duration(s.config.RefreshTokenTTL) * time.Second)),
		Issuer:    s.config.Issuer,
		Audience:  []string{s.config.Audience},
	}
	
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	token.Header["kid"] = s.keyID
	
	tokenString, err := token.SignedString(s.privateKey)
	if err != nil {
		return "", "", fmt.Errorf("failed to sign refresh token: %w", err)
	}
	
	return tokenString, jti, nil
}

func (s *jwtService) ValidateToken(tokenString string) (*TokenClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &TokenClaims{}, func(token *jwt.Token) (interface{}, error) {
		// Verify signing method
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		
		// Verify key ID
		kid, ok := token.Header["kid"].(string)
		if !ok || kid != s.keyID {
			return nil, fmt.Errorf("invalid key ID: %v", kid)
		}
		
		return s.publicKey, nil
	})
	
	if err != nil {
		// Handle specific JWT errors
		switch {
		case err.Error() == "token is expired":
			return nil, errors.ErrTokenExpired
		case err.Error() == "signature is invalid":
			return nil, errors.ErrInvalidSignature
		case err.Error() == "token is malformed":
			return nil, errors.ErrTokenInvalid.WithDetails("malformed token")
		default:
			return nil, errors.ErrTokenInvalid.WithDetails(err.Error())
		}
		return nil, errors.ErrTokenInvalid.WithDetails(err.Error())
	}
	
	if !token.Valid {
		return nil, errors.ErrTokenInvalid
	}
	
	claims, ok := token.Claims.(*TokenClaims)
	if !ok {
		return nil, errors.ErrTokenInvalid.WithDetails("invalid claims type")
	}
	
	// Additional validation
	if claims.Issuer != s.config.Issuer {
		return nil, errors.ErrTokenInvalid.WithDetails("invalid issuer")
	}
	
	if claims.Audience != s.config.Audience {
		return nil, errors.ErrTokenInvalid.WithDetails("invalid audience")
	}
	
	return claims, nil
}

func (s *jwtService) GetPublicKey() *rsa.PublicKey {
	return s.publicKey
}

func (s *jwtService) GetKeyID() string {
	return s.keyID
}

func (s *jwtService) RotateKeys() error {
	// This would typically involve:
	// 1. Generate new key pair
	// 2. Store in HSM
	// 3. Update key ID
	// 4. Keep old key for validation during transition period
	// For now, this is a placeholder
	return fmt.Errorf("key rotation not implemented - requires HSM integration")
}

// JWKSResponse represents the JWKS response format
type JWKSResponse struct {
	Keys []JWK `json:"keys"`
}

// JWK represents a JSON Web Key
type JWK struct {
	KeyType   string `json:"kty"`
	KeyID     string `json:"kid"`
	Use       string `json:"use"`
	Algorithm string `json:"alg,omitempty"`
	N         string `json:"n"`
	E         string `json:"e"`
}

// GenerateJWKS generates a JWKS response from the current public key
func (s *jwtService) GenerateJWKS() (*JWKSResponse, error) {
	// Convert RSA public key to JWK format
	n := s.publicKey.N.Bytes()
	e := big.NewInt(int64(s.publicKey.E)).Bytes()
	
	// Base64url encode without padding
	nEncoded := base64.RawURLEncoding.EncodeToString(n)
	eEncoded := base64.RawURLEncoding.EncodeToString(e)
	
	jwk := JWK{
		KeyType:   "RSA",
		KeyID:     s.keyID,
		Use:       "sig",
		Algorithm: "RS256",
		N:         nEncoded,
		E:         eEncoded,
	}
	
	return &JWKSResponse{
		Keys: []JWK{jwk},
	}, nil
}

// TokenValidator provides token validation functionality
type TokenValidator struct {
	jwtService JWTService
}

func NewTokenValidator(jwtService JWTService) *TokenValidator {
	return &TokenValidator{jwtService: jwtService}
}

func (v *TokenValidator) ValidateAndExtractClaims(tokenString string) (*TokenClaims, error) {
	// Remove "Bearer " prefix if present
	if len(tokenString) > 7 && tokenString[:7] == "Bearer " {
		tokenString = tokenString[7:]
	}
	
	return v.jwtService.ValidateToken(tokenString)
}

// RefreshTokenManager handles refresh token operations
type RefreshTokenManager struct {
	jwtService JWTService
	// tokenRepo  models.TokenRepository // This would be injected
}

func NewRefreshTokenManager(jwtService JWTService) *RefreshTokenManager {
	return &RefreshTokenManager{
		jwtService: jwtService,
	}
}

func (m *RefreshTokenManager) CreateRefreshToken(userID, tenantID, unitID string) (*models.RefreshToken, error) {
	tokenString, jti, err := m.jwtService.GenerateRefreshToken()
	if err != nil {
		return nil, err
	}
	
	// Hash the token for storage
	tokenHash, err := hashToken(tokenString)
	if err != nil {
		return nil, err
	}
	
	refreshToken := &models.RefreshToken{
		ID:        uuid.New().String(),
		JTI:       jti,
		UserID:    userID,
		TenantID:  &tenantID,
		UnitID:    &unitID,
		TokenHash: tokenHash,
		ExpiresAt: time.Now().Add(time.Duration(m.jwtService.(*jwtService).config.RefreshTokenTTL) * time.Second),
		Revoked:   false,
	}
	
	return refreshToken, nil
}

// Helper function to hash tokens for secure storage
func hashToken(token string) (string, error) {
	// Use a secure hash function (this is a simplified version)
	// In production, use proper cryptographic hashing
	return fmt.Sprintf("hashed_%s", token[:10]), nil
}

// GenerateJWKSFromService generates JWKS from a JWT service
func GenerateJWKSFromService(service JWTService) (*JWKSResponse, error) {
	if jwtSvc, ok := service.(*jwtService); ok {
		return jwtSvc.GenerateJWKS()
	}
	return nil, fmt.Errorf("unsupported JWT service type")
}