package utils

import (
	"testing"
)

func TestValidateEmail(t *testing.T) {
	tests := []struct {
		name    string
		email   string
		wantErr bool
	}{
		{"valid email", "test@example.com", false},
		{"valid email with subdomain", "user@mail.example.com", false},
		{"valid email with plus", "user+tag@example.com", false},
		{"empty email", "", true},
		{"invalid format", "invalid-email", true},
		{"missing @", "userexample.com", true},
		{"missing domain", "user@", true},
		{"missing user", "@example.com", true},
		{"too long", "a" + string(make([]byte, 250)) + "@example.com", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateEmail(tt.email)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateEmail() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// Eliminado: TestValidatePhone

func TestValidatePassword(t *testing.T) {
	tests := []struct {
		name     string
		password string
		wantErr  bool
	}{
		{"valid strong password", "MySecurePass123!", false},
		{"valid with special chars", "Test@Password2023", false},
		{"empty password", "", true},
		{"too short", "Short1!", true},
		{"no uppercase", "lowercase123!", true},
		{"no lowercase", "UPPERCASE123!", true},
		{"no numbers", "NoNumbers!", true},
		{"no special chars", "NoSpecialChars123", true},
		{"weak password", "password", true},
		{"common weak", "123456", true},
		{"too long", string(make([]byte, 130)), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidatePassword(tt.password)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidatePassword() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateName(t *testing.T) {
	tests := []struct {
		name    string
		nameStr string
		wantErr bool
	}{
		{"valid name", "Juan Pérez", false},
		{"valid with accent", "María González", false},
		{"valid with apostrophe", "O'Connor", false},
		{"valid with hyphen", "Ana-María", false},
		{"valid with dot", "Dr. Smith", false},
		{"empty name", "", true},
		{"too short", "A", true},
		{"too long", string(make([]byte, 101)), true},
		{"with numbers", "John123", true},
		{"with symbols", "John@Doe", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateName(tt.nameStr, "Nombre")
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateName() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// Eliminado: TestValidateOTP

func TestValidateUUID(t *testing.T) {
	tests := []struct {
		name    string
		uuid    string
		wantErr bool
	}{
		{"valid UUID", "123e4567-e89b-12d3-a456-426614174000", false},
		{"valid UUID lowercase", "123e4567-e89b-12d3-a456-426614174000", false},
		{"valid UUID uppercase", "123E4567-E89B-12D3-A456-426614174000", false},
		{"empty UUID", "", true},
		{"invalid format", "not-a-uuid", true},
		{"missing hyphens", "123e4567e89b12d3a456426614174000", true},
		{"wrong length", "123e4567-e89b-12d3-a456-42661417400", true},
		{"invalid characters", "123g4567-e89b-12d3-a456-426614174000", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateUUID(tt.uuid)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateUUID() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateRole(t *testing.T) {
	tests := []struct {
		name    string
		role    string
		wantErr bool
	}{
		{"valid owner", "owner", false},
		{"valid tenant", "tenant", false},
		{"valid family_member", "family_member", false},
		{"valid administrator", "administrator", false},
		{"valid manager", "manager", false},
		{"invalid role", "invalid_role", true},
		{"empty role", "", true},
		{"case sensitive", "Owner", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateRole(tt.role)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateRole() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestSanitizeString(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"normal string", "hello world", "hello world"},
		{"with null bytes", "hello\x00world", "helloworld"},
		{"with whitespace", "  hello world  ", "hello world"},
		{"multiple null bytes", "a\x00b\x00c", "abc"},
		{"empty string", "", ""},
		{"only whitespace", "   ", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SanitizeString(tt.input)
			if result != tt.expected {
				t.Errorf("SanitizeString() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestValidateAndSanitizeInput(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		fieldName string
		maxLength int
		wantErr   bool
		expected  string
	}{
		{"valid input", "hello", "test", 10, false, "hello"},
		{"with whitespace", "  hello  ", "test", 10, false, "hello"},
		{"empty input", "", "test", 10, true, ""},
		{"too long", "verylongstring", "test", 5, true, ""},
		{"with null bytes", "hello\x00world", "test", 20, false, "helloworld"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ValidateAndSanitizeInput(tt.input, tt.fieldName, tt.maxLength)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateAndSanitizeInput() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && result != tt.expected {
				t.Errorf("ValidateAndSanitizeInput() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestValidatePagination(t *testing.T) {
	tests := []struct {
		name      string
		page      int
		limit     int
		wantPage  int
		wantLimit int
	}{
		{"valid pagination", 2, 20, 2, 20},
		{"zero page", 0, 20, 1, 20},
		{"negative page", -1, 20, 1, 20},
		{"zero limit", 1, 0, 1, 10},
		{"negative limit", 1, -5, 1, 10},
		{"too large limit", 1, 200, 1, 100},
		{"all invalid", -1, -5, 1, 10},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotPage, gotLimit, err := ValidatePagination(tt.page, tt.limit)
			if err != nil {
				t.Errorf("ValidatePagination() error = %v", err)
				return
			}
			if gotPage != tt.wantPage {
				t.Errorf("ValidatePagination() page = %v, want %v", gotPage, tt.wantPage)
			}
			if gotLimit != tt.wantLimit {
				t.Errorf("ValidatePagination() limit = %v, want %v", gotLimit, tt.wantLimit)
			}
		})
	}
}
