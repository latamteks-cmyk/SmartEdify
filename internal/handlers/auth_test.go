package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/smartedify/auth-service/internal/models"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/valyala/fasthttp"
)

// MockAuthService is a mock implementation of AuthService
type MockAuthService struct {
	mock.Mock
}

func (m *MockAuthService) RegisterUser(ctx *fasthttp.RequestCtx, req *models.CreateUserRequest) (*models.User, error) {
	args := m.Called(ctx, req)
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockAuthService) GetUserByID(ctx *fasthttp.RequestCtx, userID string) (*models.User, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockAuthService) UpdateUserStatus(ctx *fasthttp.RequestCtx, userID, status string) error {
	args := m.Called(ctx, userID, status)
	return args.Error(0)
}

func (m *MockAuthService) LoginWithWhatsApp(ctx *fasthttp.RequestCtx, req *models.WhatsAppLoginRequest) (*models.LoginResponse, error) {
	args := m.Called(ctx, req)
	return args.Get(0).(*models.LoginResponse), args.Error(1)
}

func (m *MockAuthService) LoginWithPassword(ctx *fasthttp.RequestCtx, req *models.LoginRequest) (*models.LoginResponse, error) {
	args := m.Called(ctx, req)
	return args.Get(0).(*models.LoginResponse), args.Error(1)
}

func (m *MockAuthService) SendWhatsAppOTP(ctx *fasthttp.RequestCtx, phone string) error {
	args := m.Called(ctx, phone)
	return args.Error(0)
}

func (m *MockAuthService) RefreshToken(ctx *fasthttp.RequestCtx, req *models.RefreshTokenRequest) (*models.LoginResponse, error) {
	args := m.Called(ctx, req)
	return args.Get(0).(*models.LoginResponse), args.Error(1)
}

func (m *MockAuthService) RevokeToken(ctx *fasthttp.RequestCtx, jti string) error {
	args := m.Called(ctx, jti)
	return args.Error(0)
}

func (m *MockAuthService) RevokeAllUserTokens(ctx *fasthttp.RequestCtx, userID string) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}

func (m *MockAuthService) GetUserPermissions(ctx *fasthttp.RequestCtx, userID, tenantID, unitID string) (*models.UserPermissions, error) {
	args := m.Called(ctx, userID, tenantID, unitID)
	return args.Get(0).(*models.UserPermissions), args.Error(1)
}

func (m *MockAuthService) TransferPresident(ctx *fasthttp.RequestCtx, tenantID, toUserID, fromUserID string) error {
	args := m.Called(ctx, tenantID, toUserID, fromUserID)
	return args.Error(0)
}

func (m *MockAuthService) GetTenantPresident(ctx *fasthttp.RequestCtx, tenantID string) (*models.TenantPresident, error) {
	args := m.Called(ctx, tenantID)
	return args.Get(0).(*models.TenantPresident), args.Error(1)
}

func TestAuthHandler_SendWhatsAppOTP(t *testing.T) {
	// Setup
	mockService := new(MockAuthService)
	handler := NewAuthHandler(mockService, nil)

	app := fiber.New()
	app.Post("/otp/send", handler.SendWhatsAppOTP)

	tests := []struct {
		name           string
		requestBody    map[string]interface{}
		mockSetup      func()
		expectedStatus int
	}{
		{
			name: "successful OTP send",
			requestBody: map[string]interface{}{
				"phone": "+51987654321",
			},
			mockSetup:      func() {},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "invalid request body",
			requestBody:    nil, // Usaremos un body inválido manualmente
			mockSetup:      func() {},
			expectedStatus: http.StatusBadRequest,
		},
	}

	// Configuración global del mock para todas las llamadas
	mockService.On("SendWhatsAppOTP", mock.Anything, mock.Anything).Return(nil)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// No configurar el mock individualmente

			var req *http.Request
			if tt.name == "invalid request body" {
				// Body inválido: un string en vez de JSON
				req = httptest.NewRequest("POST", "/otp/send", bytes.NewBufferString("no-json"))
				req.Header.Set("Content-Type", "application/json")
			} else {
				body, _ := json.Marshal(tt.requestBody)
				req = httptest.NewRequest("POST", "/otp/send", bytes.NewBuffer(body))
				req.Header.Set("Content-Type", "application/json")
			}

			resp, err := app.Test(req)
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedStatus, resp.StatusCode)

			mockService.ExpectedCalls = nil
		})
	}
}

func TestAuthHandler_LoginWithWhatsApp(t *testing.T) {
	// Setup
	mockService := new(MockAuthService)
	handler := NewAuthHandler(mockService, nil)

	app := fiber.New()
	app.Post("/login/whatsapp", handler.LoginWithWhatsApp)

	tests := []struct {
		name           string
		requestBody    models.WhatsAppLoginRequest
		mockSetup      func()
		expectedStatus int
	}{
		{
			name: "successful login",
			requestBody: models.WhatsAppLoginRequest{
				Phone: "+51987654321",
				OTP:   "123456",
			},
			mockSetup: func() {
				loginResponse := &models.LoginResponse{
					AccessToken:  "access_token",
					RefreshToken: "refresh_token",
					ExpiresIn:    3600,
					TokenType:    "Bearer",
					UserID:       "user_id",
				}
				mockService.On("LoginWithWhatsApp", mock.Anything, mock.AnythingOfType("*models.WhatsAppLoginRequest")).Return(loginResponse, nil)
			},
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mock
			tt.mockSetup()

			// Create request
			body, _ := json.Marshal(tt.requestBody)
			req := httptest.NewRequest("POST", "/login/whatsapp", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")

			// Execute request
			resp, err := app.Test(req)
			assert.NoError(t, err)

			// Assert
			assert.Equal(t, tt.expectedStatus, resp.StatusCode)

			// Clear mock expectations
			mockService.ExpectedCalls = nil
		})
	}
}

func TestAuthHandler_Register(t *testing.T) {
	// Setup
	mockService := new(MockAuthService)
	handler := NewAuthHandler(mockService, nil)

	app := fiber.New()
	app.Post("/register", handler.Register)

	tests := []struct {
		name           string
		requestBody    models.CreateUserRequest
		mockSetup      func()
		expectedStatus int
	}{
		{
			name: "successful registration",
			requestBody: models.CreateUserRequest{
				Email:         "test@example.com",
				Phone:         "+51987654321",
				Name:          "Test User",
				TenantID:      "tenant_id",
				TermsAccepted: true,
			},
			mockSetup: func() {
				// Corregir el tipo de ID en el test para usar uuid.UUID
				user := &models.User{
					ID:    uuid.MustParse("11111111-1111-1111-1111-111111111111"),
					Email: "test@example.com",
					Phone: "+51987654321",
					Name:  "Test User",
				}
				mockService.On("RegisterUser", mock.Anything, mock.AnythingOfType("*models.CreateUserRequest")).Return(user, nil)
			},
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mock
			tt.mockSetup()

			// Create request
			body, _ := json.Marshal(tt.requestBody)
			req := httptest.NewRequest("POST", "/register", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")

			// Execute request
			resp, err := app.Test(req)
			assert.NoError(t, err)

			// Assert
			assert.Equal(t, tt.expectedStatus, resp.StatusCode)

			// Clear mock expectations
			mockService.ExpectedCalls = nil
		})
	}
}
