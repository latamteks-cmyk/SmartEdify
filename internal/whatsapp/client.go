package whatsapp

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/smartedify/auth-service/internal/config"
	"github.com/smartedify/auth-service/internal/errors"
)

// WhatsAppClient interface for WhatsApp API operations
type WhatsAppClient interface {
	SendOTP(ctx context.Context, phone, otp string) error
	SendMessage(ctx context.Context, phone, message string) error
	SendTemplateMessage(ctx context.Context, phone, templateName string, params map[string]string) error
}

// whatsappClient implements WhatsApp Business API client
type whatsappClient struct {
	config     *config.WhatsAppConfig
	httpClient *http.Client
}

// NewWhatsAppClient creates a new WhatsApp client
func NewWhatsAppClient(cfg *config.WhatsAppConfig) WhatsAppClient {
	return &whatsappClient{
		config: cfg,
		httpClient: &http.Client{
			Timeout: time.Duration(cfg.Timeout) * time.Second,
		},
	}
}

// SendOTPRequest represents the request to send OTP
type SendOTPRequest struct {
	To       string `json:"to"`
	Type     string `json:"type"`
	Template struct {
		Name       string `json:"name"`
		Language   struct {
			Code string `json:"code"`
		} `json:"language"`
		Components []struct {
			Type       string `json:"type"`
			Parameters []struct {
				Type string `json:"type"`
				Text string `json:"text"`
			} `json:"parameters"`
		} `json:"components"`
	} `json:"template"`
}

// SendMessageRequest represents a simple message request
type SendMessageRequest struct {
	To   string `json:"to"`
	Type string `json:"type"`
	Text struct {
		Body string `json:"body"`
	} `json:"text"`
}

// WhatsAppResponse represents the API response
type WhatsAppResponse struct {
	Messages []struct {
		ID string `json:"id"`
	} `json:"messages"`
	Error struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
		Details string `json:"details"`
	} `json:"error,omitempty"`
}

func (w *whatsappClient) SendOTP(ctx context.Context, phone, otp string) error {
	// Format phone number (remove + if present)
	if phone[0] == '+' {
		phone = phone[1:]
	}
	
	// Create OTP message using template
	request := SendOTPRequest{
		To:   phone,
		Type: "template",
	}
	
	request.Template.Name = "smartedify_otp"
	request.Template.Language.Code = "es"
	request.Template.Components = []struct {
		Type       string `json:"type"`
		Parameters []struct {
			Type string `json:"type"`
			Text string `json:"text"`
		} `json:"parameters"`
	}{
		{
			Type: "body",
			Parameters: []struct {
				Type string `json:"type"`
				Text string `json:"text"`
			}{
				{
					Type: "text",
					Text: otp,
				},
			},
		},
	}
	
	return w.sendRequest(ctx, request)
}

func (w *whatsappClient) SendMessage(ctx context.Context, phone, message string) error {
	// Format phone number
	if phone[0] == '+' {
		phone = phone[1:]
	}
	
	request := SendMessageRequest{
		To:   phone,
		Type: "text",
	}
	request.Text.Body = message
	
	return w.sendRequest(ctx, request)
}

func (w *whatsappClient) SendTemplateMessage(ctx context.Context, phone, templateName string, params map[string]string) error {
	// Format phone number
	if phone[0] == '+' {
		phone = phone[1:]
	}
	
	request := SendOTPRequest{
		To:   phone,
		Type: "template",
	}
	
	request.Template.Name = templateName
	request.Template.Language.Code = "es"
	
	// Convert params to template parameters
	var parameters []struct {
		Type string `json:"type"`
		Text string `json:"text"`
	}
	
	for _, value := range params {
		parameters = append(parameters, struct {
			Type string `json:"type"`
			Text string `json:"text"`
		}{
			Type: "text",
			Text: value,
		})
	}
	
	request.Template.Components = []struct {
		Type       string `json:"type"`
		Parameters []struct {
			Type string `json:"type"`
			Text string `json:"text"`
		} `json:"parameters"`
	}{
		{
			Type:       "body",
			Parameters: parameters,
		},
	}
	
	return w.sendRequest(ctx, request)
}

func (w *whatsappClient) sendRequest(ctx context.Context, payload interface{}) error {
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}
	
	req, err := http.NewRequestWithContext(ctx, "POST", w.config.APIEndpoint, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	
	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+w.config.APIKey)
	
	resp, err := w.httpClient.Do(req)
	if err != nil {
		return errors.ErrWhatsAppFailed.WithDetails(err.Error())
	}
	defer resp.Body.Close()
	
	var response WhatsAppResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}
	
	if resp.StatusCode != http.StatusOK {
		return errors.ErrWhatsAppFailed.WithDetails(fmt.Sprintf("API error: %s", response.Error.Message))
	}
	
	if response.Error.Code != 0 {
		return errors.ErrWhatsAppFailed.WithDetails(response.Error.Message)
	}
	
	return nil
}

// MockWhatsAppClient for testing and development
type MockWhatsAppClient struct {
	sentMessages []MockMessage
}

type MockMessage struct {
	Phone     string
	Message   string
	Template  string
	Timestamp time.Time
}

func NewMockWhatsAppClient() *MockWhatsAppClient {
	return &MockWhatsAppClient{
		sentMessages: make([]MockMessage, 0),
	}
}

func (m *MockWhatsAppClient) SendOTP(ctx context.Context, phone, otp string) error {
	message := fmt.Sprintf("Tu código de verificación SmartEdify es: %s. Válido por 5 minutos.", otp)
	m.sentMessages = append(m.sentMessages, MockMessage{
		Phone:     phone,
		Message:   message,
		Template:  "smartedify_otp",
		Timestamp: time.Now(),
	})
	
	// Simulate API delay
	time.Sleep(100 * time.Millisecond)
	
	return nil
}

func (m *MockWhatsAppClient) SendMessage(ctx context.Context, phone, message string) error {
	m.sentMessages = append(m.sentMessages, MockMessage{
		Phone:     phone,
		Message:   message,
		Timestamp: time.Now(),
	})
	
	time.Sleep(100 * time.Millisecond)
	return nil
}

func (m *MockWhatsAppClient) SendTemplateMessage(ctx context.Context, phone, templateName string, params map[string]string) error {
	message := fmt.Sprintf("Template: %s with params: %v", templateName, params)
	m.sentMessages = append(m.sentMessages, MockMessage{
		Phone:     phone,
		Message:   message,
		Template:  templateName,
		Timestamp: time.Now(),
	})
	
	time.Sleep(100 * time.Millisecond)
	return nil
}

func (m *MockWhatsAppClient) GetSentMessages() []MockMessage {
	return m.sentMessages
}

func (m *MockWhatsAppClient) ClearMessages() {
	m.sentMessages = make([]MockMessage, 0)
}

// WhatsAppService provides high-level WhatsApp operations
type WhatsAppService struct {
	client WhatsAppClient
}

func NewWhatsAppService(client WhatsAppClient) *WhatsAppService {
	return &WhatsAppService{client: client}
}

func (s *WhatsAppService) SendLoginOTP(ctx context.Context, phone, otp string) error {
	return s.client.SendOTP(ctx, phone, otp)
}

func (s *WhatsAppService) SendPresidentInvitation(ctx context.Context, phone, presidentName, tenantName string) error {
	message := fmt.Sprintf(
		"Hola %s, has sido designado como presidente de %s. Para aceptar el cargo, responde 'SÍ' a este mensaje.",
		presidentName, tenantName,
	)
	return s.client.SendMessage(ctx, phone, message)
}

func (s *WhatsAppService) SendAssemblyNotification(ctx context.Context, phone, assemblyTitle string, assemblyDate time.Time) error {
	message := fmt.Sprintf(
		"Convocatoria: %s programada para el %s. Más detalles en tu app SmartEdify.",
		assemblyTitle, assemblyDate.Format("02/01/2006 15:04"),
	)
	return s.client.SendMessage(ctx, phone, message)
}

func (s *WhatsAppService) SendPaymentReminder(ctx context.Context, phone, amount, dueDate string) error {
	message := fmt.Sprintf(
		"Recordatorio: Tienes una cuota pendiente de %s con vencimiento el %s. Paga desde tu app SmartEdify.",
		amount, dueDate,
	)
	return s.client.SendMessage(ctx, phone, message)
}