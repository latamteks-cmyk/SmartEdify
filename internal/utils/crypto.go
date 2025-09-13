package utils

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"math/big"
	"strings"

	"golang.org/x/crypto/bcrypt"
	"golang.org/x/crypto/argon2"
)

const (
	// bcrypt cost factor - aligned with spec requirements (cost factor 12)
	bcryptCost = 12
	
	// Argon2id parameters for OTP hashing
	argon2Time    = 1
	argon2Memory  = 32 * 1024
	argon2Threads = 2
	argon2KeyLen  = 32
	saltLength    = 16
)

// GenerateRandomString generates a cryptographically secure random string
func GenerateRandomString(length int) (string, error) {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	for i := range b {
		num, err := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		if err != nil {
			return "", err
		}
		b[i] = charset[num.Int64()]
	}
	return string(b), nil
}

// GenerateRandomBytes generates cryptographically secure random bytes
func GenerateRandomBytes(length int) ([]byte, error) {
	bytes := make([]byte, length)
	_, err := rand.Read(bytes)
	if err != nil {
		return nil, err
	}
	return bytes, nil
}

// GenerateOTP generates a numeric OTP code
func GenerateOTP(length int) (string, error) {
	if length < 4 || length > 8 {
		length = 6 // Default to 6 digits
	}
	
	max := int64(1)
	for i := 0; i < length; i++ {
		max *= 10
	}
	
	num, err := rand.Int(rand.Reader, big.NewInt(max))
	if err != nil {
		return "", err
	}
	
	return fmt.Sprintf("%0*d", length, num.Int64()), nil
}

// HashPassword hashes a password using bcrypt - aligned with spec requirements (cost factor 12)
func HashPassword(password string) (string, error) {
	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(password), bcryptCost)
	if err != nil {
		return "", fmt.Errorf("failed to hash password: %w", err)
	}
	
	return string(hashedBytes), nil
}

// VerifyPassword verifies a password against its bcrypt hash - aligned with spec requirements
func VerifyPassword(password, hashedPassword string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	return err == nil
}

// HashPasswordArgon2 hashes a password using Argon2id (alternative method)
func HashPasswordArgon2(password string) (string, error) {
	salt, err := GenerateRandomBytes(saltLength)
	if err != nil {
		return "", err
	}
	
	hash := argon2.IDKey([]byte(password), salt, argon2Time, argon2Memory, argon2Threads, argon2KeyLen)
	
	// Encode as: $argon2id$v=19$m=32768,t=1,p=2$salt$hash
	encodedSalt := base64.RawStdEncoding.EncodeToString(salt)
	encodedHash := base64.RawStdEncoding.EncodeToString(hash)
	
	return fmt.Sprintf("$argon2id$v=19$m=%d,t=%d,p=%d$%s$%s",
		argon2Memory, argon2Time, argon2Threads, encodedSalt, encodedHash), nil
}

// VerifyPasswordArgon2 verifies a password against its Argon2id hash (alternative method)
func VerifyPasswordArgon2(password, hashedPassword string) bool {
	parts := strings.Split(hashedPassword, "$")
	if len(parts) != 6 {
		return false
	}
	
	if parts[1] != "argon2id" || parts[2] != "v=19" {
		return false
	}
	
	var memory, time, threads uint32
	_, err := fmt.Sscanf(parts[3], "m=%d,t=%d,p=%d", &memory, &time, &threads)
	if err != nil {
		return false
	}
	
	salt, err := base64.RawStdEncoding.DecodeString(parts[4])
	if err != nil {
		return false
	}
	
	expectedHash, err := base64.RawStdEncoding.DecodeString(parts[5])
	if err != nil {
		return false
	}
	
	actualHash := argon2.IDKey([]byte(password), salt, time, memory, uint8(threads), uint32(len(expectedHash)))
	
	return subtle.ConstantTimeCompare(expectedHash, actualHash) == 1
}

// HashOTP hashes an OTP code for secure storage
func HashOTP(otp string) (string, error) {
	salt, err := GenerateRandomBytes(16)
	if err != nil {
		return "", err
	}
	
	hash := argon2.IDKey([]byte(otp), salt, 1, 32*1024, 2, 32)
	
	encodedSalt := base64.RawStdEncoding.EncodeToString(salt)
	encodedHash := base64.RawStdEncoding.EncodeToString(hash)
	
	return fmt.Sprintf("%s$%s", encodedSalt, encodedHash), nil
}

// VerifyOTP verifies an OTP against its hash
func VerifyOTP(otp, hashedOTP string) bool {
	parts := strings.Split(hashedOTP, "$")
	if len(parts) != 2 {
		return false
	}
	
	salt, err := base64.RawStdEncoding.DecodeString(parts[0])
	if err != nil {
		return false
	}
	
	expectedHash, err := base64.RawStdEncoding.DecodeString(parts[1])
	if err != nil {
		return false
	}
	
	actualHash := argon2.IDKey([]byte(otp), salt, 1, 32*1024, 2, 32)
	
	return subtle.ConstantTimeCompare(expectedHash, actualHash) == 1
}

// GenerateSecureToken generates a secure token for various purposes
func GenerateSecureToken() (string, error) {
	bytes, err := GenerateRandomBytes(32)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// GenerateJTI generates a unique JWT ID
func GenerateJTI() (string, error) {
	bytes, err := GenerateRandomBytes(16)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// GenerateNonce generates a nonce for DPoP
func GenerateNonce() (string, error) {
	bytes, err := GenerateRandomBytes(16)
	if err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(bytes), nil
}