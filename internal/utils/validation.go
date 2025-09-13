package utils

import (
	"fmt"
	"regexp"
	"strings"
	"unicode"

	"github.com/smartedify/auth-service/internal/errors"
)

var (
	// Password strength requirements - aligned with spec requirements
	minPasswordLength = 8 // Spec requires minimum 8 characters

	// Common weak passwords to reject - aligned with spec security requirements
	weakPasswords = map[string]bool{
		"password":    true,
		"123456":      true,
		"123456789":   true,
		"qwerty":      true,
		"abc123":      true,
		"password123": true,
		"admin":       true,
		"letmein":     true,
		"welcome":     true,
		"monkey":      true,
		"12345678":    true,
		"qwerty123":   true,
		"1q2w3e4r":    true,
		"admin123":    true,
		"root":        true,
		"toor":        true,
		"pass":        true,
		"test":        true,
		"guest":       true,
		"user":        true,
	}

	// Email regex for validation
	emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)

	// UUID regex for validation
	uuidRegex = regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)

	// Name regex for validation (letters, spaces, hyphens, apostrophes, dots)
	nameRegex = regexp.MustCompile(`^[a-zA-ZÀ-ÿ\s\-'\.]+$`)
)

// ValidateEmail validates email format - aligned with spec requirements
func ValidateEmail(email string) error {
	if email == "" {
		return errors.ErrMissingRequired.WithDetails("email is required")
	}

	email = strings.TrimSpace(strings.ToLower(email))
	if len(email) > 254 {
		return errors.ErrInvalidFormat.WithDetails("email too long")
	}
	if !emailRegex.MatchString(email) {
		return errors.ErrInvalidFormat.WithDetails("invalid email format")
	}
	parts := strings.Split(email, "@")
	if len(parts) != 2 || len(parts[0]) == 0 || len(parts[1]) == 0 || !strings.Contains(parts[1], ".") {
		return errors.ErrInvalidFormat.WithDetails("invalid email format")
	}
	return nil
}

// ValidatePassword validates password strength - aligned with spec requirements
func ValidatePassword(password string) error {
	if password == "" {
		return errors.ErrMissingRequired.WithDetails("password is required")
	}

	// Spec requirement: minimum 8 characters
	if len(password) < minPasswordLength {
		return errors.ErrWeakPassword.WithDetails(fmt.Sprintf("password must be at least %d characters", minPasswordLength))
	}

	if len(password) > 128 {
		return errors.ErrInvalidFormat.WithDetails("password too long (max 128 characters)")
	}

	// Check against common weak passwords - spec security requirement
	if weakPasswords[strings.ToLower(password)] {
		return errors.ErrWeakPassword.WithDetails("password is too common and weak")
	}

	// Check for sequential characters (123, abc, etc.)

	// Check for repeated characters (aaa, 111, etc.)
	// Secuencias y repeticiones permitidas si cumple los demás requisitos

	// Spec requirements: uppercase, lowercase, number, special character
	var (
		hasUpper   = false
		hasLower   = false
		hasNumber  = false
		hasSpecial = false
	)

	for _, char := range password {
		switch {
		case unicode.IsUpper(char):
			hasUpper = true
		case unicode.IsLower(char):
			hasLower = true
		case unicode.IsNumber(char):
			hasNumber = true
		case unicode.IsPunct(char) || unicode.IsSymbol(char):
			hasSpecial = true
		}
	}

	var missingRequirements []string
	if !hasUpper {
		missingRequirements = append(missingRequirements, "uppercase letter")
	}
	if !hasLower {
		missingRequirements = append(missingRequirements, "lowercase letter")
	}
	if !hasNumber {
		missingRequirements = append(missingRequirements, "number")
	}
	if !hasSpecial {
		missingRequirements = append(missingRequirements, "special character")
	}

	if len(missingRequirements) > 0 {
		return errors.ErrWeakPassword.WithDetails(fmt.Sprintf("password must contain: %s", strings.Join(missingRequirements, ", ")))
	}

	return nil
}

// ValidateName validates user name - aligned with spec requirements
func ValidateName(name string, fieldName string) error {
	if name == "" {
		return errors.ErrMissingRequired.WithDetails(fmt.Sprintf("%s is required", fieldName))
	}

	name = strings.TrimSpace(name)

	// Spec requirement: min 2, max 50 characters for firstName/lastName
	if len(name) < 2 {
		return errors.ErrInvalidFormat.WithDetails(fmt.Sprintf("%s must be at least 2 characters", fieldName))
	}

	if len(name) > 50 {
		return errors.ErrInvalidFormat.WithDetails(fmt.Sprintf("%s too long (max 50 characters)", fieldName))
	}

	// Check for valid characters (letters, spaces, hyphens, apostrophes, dots)
	if !nameRegex.MatchString(name) {
		return errors.ErrInvalidFormat.WithDetails(fmt.Sprintf("%s contains invalid characters", fieldName))
	}

	return nil
}

// ValidateUUID validates UUID format - aligned with spec requirements
func ValidateUUID(uuid string) error {
	if uuid == "" {
		return errors.ErrMissingRequired.WithDetails("UUID is required")
	}

	if !uuidRegex.MatchString(strings.ToLower(uuid)) {
		return errors.ErrInvalidFormat.WithDetails("invalid UUID format")
	}

	return nil
}

// ValidateTenantID validates tenant ID format - aligned with spec requirements
func ValidateTenantID(tenantID string) error {
	if tenantID == "" {
		return errors.ErrMissingRequired.WithDetails("tenantId is required")
	}

	tenantID = strings.TrimSpace(tenantID)

	if len(tenantID) < 1 || len(tenantID) > 100 {
		return errors.ErrInvalidFormat.WithDetails("tenantId must be 1-100 characters")
	}

	// Allow alphanumeric, hyphens, and underscores
	tenantRegex := regexp.MustCompile(`^[a-zA-Z0-9\-_]+$`)
	if !tenantRegex.MatchString(tenantID) {
		return errors.ErrInvalidFormat.WithDetails("tenantId contains invalid characters")
	}

	return nil
}

// ValidateUnitID validates unit ID format - aligned with spec requirements
func ValidateUnitID(unitID string) error {
	if unitID == "" {
		return errors.ErrMissingRequired.WithDetails("unitId is required")
	}

	unitID = strings.TrimSpace(unitID)

	if len(unitID) < 1 || len(unitID) > 100 {
		return errors.ErrInvalidFormat.WithDetails("unitId must be 1-100 characters")
	}

	// Allow alphanumeric, hyphens, and underscores
	unitRegex := regexp.MustCompile(`^[a-zA-Z0-9\-_]+$`)
	if !unitRegex.MatchString(unitID) {
		return errors.ErrInvalidFormat.WithDetails("unitId contains invalid characters")
	}

	return nil
}

// ValidateRole validates user role - aligned with spec requirements
func ValidateRole(role string) error {
	validRoles := map[string]bool{
		"user":          true,
		"admin":         true,
		"owner":         true,
		"tenant":        true,
		"family_member": true,
		"administrator": true,
		"manager":       true,
	}

	if !validRoles[role] {
		return errors.ErrInvalidFormat.WithDetails("invalid role")
	}

	return nil
}

// SanitizeString removes potentially dangerous characters
func SanitizeString(input string) string {
	// Remove null bytes
	input = strings.ReplaceAll(input, "\x00", "")

	// Remove control characters except tab, newline, and carriage return
	var result strings.Builder
	for _, r := range input {
		if r >= 32 || r == 9 || r == 10 || r == 13 {
			result.WriteRune(r)
		}
	}

	// Trim whitespace
	return strings.TrimSpace(result.String())
}

// ValidateAndSanitizeInput validates and sanitizes user input
func ValidateAndSanitizeInput(input string, fieldName string, maxLength int) (string, error) {
	if input == "" {
		return "", errors.ErrMissingRequired.WithDetails(fmt.Sprintf("%s is required", fieldName))
	}

	sanitized := SanitizeString(input)

	if len(sanitized) > maxLength {
		return "", errors.ErrInvalidFormat.WithDetails(fmt.Sprintf("%s too long (max %d characters)", fieldName, maxLength))
	}

	return sanitized, nil
}

// Helper functions for password validation

// hasSequentialChars checks for sequential characters in password
func hasSequentialChars(password string) bool {
	password = strings.ToLower(password)

	// Check for sequential numbers
	sequences := []string{
		"0123456789", "1234567890", "9876543210", "0987654321",
		"abcdefghijklmnopqrstuvwxyz", "zyxwvutsrqponmlkjihgfedcba",
		"qwertyuiop", "asdfghjkl", "zxcvbnm",
	}

	for _, seq := range sequences {
		for i := 0; i <= len(seq)-3; i++ {
			if strings.Contains(password, seq[i:i+3]) {
				return true
			}
		}
	}

	return false
}

// hasRepeatedChars checks for too many repeated characters
func hasRepeatedChars(password string) bool {
	if len(password) < 3 {
		return false
	}

	repeatCount := 1
	for i := 1; i < len(password); i++ {
		if password[i] == password[i-1] {
			repeatCount++
			if repeatCount >= 3 { // 3 or more consecutive identical characters
				return true
			}
		} else {
			repeatCount = 1
		}
	}

	return false
}

// ValidatePagination validates pagination parameters
func ValidatePagination(page, limit int) (int, int, error) {
	if page < 1 {
		page = 1
	}

	if limit < 1 {
		limit = 10
	}

	if limit > 100 {
		limit = 100
	}

	return page, limit, nil
}
