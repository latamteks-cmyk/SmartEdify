package service

import (
	"regexp"
	"unicode"

	"golang.org/x/crypto/bcrypt"
)

// HashPassword genera el hash de la contraseña usando bcrypt con cost factor 12
func HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

// CheckPassword compara la contraseña con el hash almacenado
func CheckPassword(hash, password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

var commonPasswords = map[string]struct{}{
	"123456": {}, "password": {}, "123456789": {}, "qwerty": {}, "abc123": {}, "password1": {},
	// ...agregar más si es necesario
}

// PasswordPolicyError describe el motivo de fallo de la validación
type PasswordPolicyError struct {
	Reason string
}

func (e *PasswordPolicyError) Error() string {
	return e.Reason
}

// ValidatePassword verifica la fortaleza de la contraseña
func ValidatePassword(password string) error {
	if len(password) < 8 {
		return &PasswordPolicyError{Reason: "La contraseña debe tener al menos 8 caracteres."}
	}

	var hasUpper, hasLower, hasNumber, hasSpecial bool
	for _, c := range password {
		switch {
		case unicode.IsUpper(c):
			hasUpper = true
		case unicode.IsLower(c):
			hasLower = true
		case unicode.IsDigit(c):
			hasNumber = true
		case regexp.MustCompile(`[!@#\$%\^&\*\-_]`).MatchString(string(c)):
			hasSpecial = true
		}
	}

	if !hasUpper {
		return &PasswordPolicyError{Reason: "Debe contener al menos una mayúscula."}
	}
	if !hasLower {
		return &PasswordPolicyError{Reason: "Debe contener al menos una minúscula."}
	}
	if !hasNumber {
		return &PasswordPolicyError{Reason: "Debe contener al menos un número."}
	}
	if !hasSpecial {
		return &PasswordPolicyError{Reason: "Debe contener al menos un carácter especial."}
	}

	if _, found := commonPasswords[password]; found {
		return &PasswordPolicyError{Reason: "La contraseña es demasiado común."}
	}

	return nil
}
