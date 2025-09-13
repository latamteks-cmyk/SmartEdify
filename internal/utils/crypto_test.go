package utils

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHashPassword(t *testing.T) {
	tests := []struct {
		name     string
		password string
		wantErr  bool
	}{
		{
			name:     "valid password",
			password: "TestPassword123!",
			wantErr:  false,
		},
		{
			name:     "another valid password",
			password: "MySecure@Pass1",
			wantErr:  false,
		},
		{
			name:     "empty password",
			password: "",
			wantErr:  false, // bcrypt can handle empty strings
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hash, err := HashPassword(tt.password)
			
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			
			require.NoError(t, err)
			assert.NotEmpty(t, hash)
			assert.NotEqual(t, tt.password, hash)
			assert.True(t, strings.HasPrefix(hash, "$2a$12$") || strings.HasPrefix(hash, "$2b$12$"))
			
			// Verify the password works
			assert.True(t, VerifyPassword(tt.password, hash))
			assert.False(t, VerifyPassword("wrongpassword", hash))
		})
	}
}

func TestVerifyPassword(t *testing.T) {
	password := "TestPassword123!"
	hash, err := HashPassword(password)
	require.NoError(t, err)

	tests := []struct {
		name     string
		password string
		hash     string
		expected bool
	}{
		{
			name:     "correct password",
			password: password,
			hash:     hash,
			expected: true,
		},
		{
			name:     "wrong password",
			password: "WrongPassword",
			hash:     hash,
			expected: false,
		},
		{
			name:     "empty password",
			password: "",
			hash:     hash,
			expected: false,
		},
		{
			name:     "invalid hash",
			password: password,
			hash:     "invalid_hash",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := VerifyPassword(tt.password, tt.hash)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGenerateOTP(t *testing.T) {
	tests := []struct {
		name   string
		length int
	}{
		{"4 digit OTP", 4},
		{"6 digit OTP", 6},
		{"8 digit OTP", 8},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			otp, err := GenerateOTP(tt.length)
			require.NoError(t, err)
			assert.Len(t, otp, tt.length)
			
			// Should be numeric
			for _, char := range otp {
				assert.True(t, char >= '0' && char <= '9', "OTP should contain only digits")
			}
			
			// Generate multiple OTPs to ensure they're different
			otp2, err := GenerateOTP(tt.length)
			require.NoError(t, err)
			
			// While it's possible they could be the same, it's extremely unlikely
			// This test might occasionally fail, but it's good for catching obvious issues
			if tt.length >= 6 {
				assert.NotEqual(t, otp, otp2, "Generated OTPs should be different")
			}
		})
	}
}

func TestHashOTP(t *testing.T) {
	otp := "123456"
	
	hash, err := HashOTP(otp)
	require.NoError(t, err)
	assert.NotEmpty(t, hash)
	assert.NotEqual(t, otp, hash)
	assert.Contains(t, hash, "$", "Hash should contain separator")
	
	// Verify the OTP
	assert.True(t, VerifyOTP(otp, hash))
	assert.False(t, VerifyOTP("654321", hash))
	assert.False(t, VerifyOTP("", hash))
}

func TestVerifyOTP(t *testing.T) {
	otp := "123456"
	hash, err := HashOTP(otp)
	require.NoError(t, err)

	tests := []struct {
		name     string
		otp      string
		hash     string
		expected bool
	}{
		{
			name:     "correct OTP",
			otp:      otp,
			hash:     hash,
			expected: true,
		},
		{
			name:     "wrong OTP",
			otp:      "654321",
			hash:     hash,
			expected: false,
		},
		{
			name:     "empty OTP",
			otp:      "",
			hash:     hash,
			expected: false,
		},
		{
			name:     "invalid hash format",
			otp:      otp,
			hash:     "invalid_hash",
			expected: false,
		},
		{
			name:     "malformed hash",
			otp:      otp,
			hash:     "part1$part2$part3",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := VerifyOTP(tt.otp, tt.hash)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGenerateRandomString(t *testing.T) {
	tests := []int{8, 16, 32, 64}

	for _, length := range tests {
		t.Run(fmt.Sprintf("length_%d", length), func(t *testing.T) {
			str, err := GenerateRandomString(length)
			require.NoError(t, err)
			assert.Len(t, str, length)
			
			// Check that it contains valid characters
			validChars := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
			for _, char := range str {
				assert.Contains(t, validChars, string(char))
			}
			
			// Generate another and ensure they're different
			str2, err := GenerateRandomString(length)
			require.NoError(t, err)
			assert.NotEqual(t, str, str2)
		})
	}
}

func TestGenerateRandomBytes(t *testing.T) {
	tests := []int{16, 32, 64}

	for _, length := range tests {
		t.Run(fmt.Sprintf("length_%d", length), func(t *testing.T) {
			bytes, err := GenerateRandomBytes(length)
			require.NoError(t, err)
			assert.Len(t, bytes, length)
			
			// Generate another and ensure they're different
			bytes2, err := GenerateRandomBytes(length)
			require.NoError(t, err)
			assert.NotEqual(t, bytes, bytes2)
		})
	}
}

func TestGenerateSecureToken(t *testing.T) {
	token, err := GenerateSecureToken()
	require.NoError(t, err)
	assert.NotEmpty(t, token)
	assert.Len(t, token, 64) // 32 bytes = 64 hex chars
	
	// Should be valid hex
	for _, char := range token {
		assert.True(t, (char >= '0' && char <= '9') || (char >= 'a' && char <= 'f'))
	}
	
	// Generate another and ensure they're different
	token2, err := GenerateSecureToken()
	require.NoError(t, err)
	assert.NotEqual(t, token, token2)
}

func TestGenerateJTI(t *testing.T) {
	jti, err := GenerateJTI()
	require.NoError(t, err)
	assert.NotEmpty(t, jti)
	assert.Len(t, jti, 32) // 16 bytes = 32 hex chars
	
	// Should be valid hex
	for _, char := range jti {
		assert.True(t, (char >= '0' && char <= '9') || (char >= 'a' && char <= 'f'))
	}
	
	// Generate another and ensure they're different
	jti2, err := GenerateJTI()
	require.NoError(t, err)
	assert.NotEqual(t, jti, jti2)
}

func TestGenerateNonce(t *testing.T) {
	nonce, err := GenerateNonce()
	require.NoError(t, err)
	assert.NotEmpty(t, nonce)
	
	// Should be valid base64 URL encoding
	// Length should be reasonable (16 bytes base64 encoded)
	assert.True(t, len(nonce) > 0)
	
	// Generate another and ensure they're different
	nonce2, err := GenerateNonce()
	require.NoError(t, err)
	assert.NotEqual(t, nonce, nonce2)
}

// Benchmark tests for performance
func BenchmarkHashPassword(b *testing.B) {
	password := "TestPassword123!"
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := HashPassword(password)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkVerifyPassword(b *testing.B) {
	password := "TestPassword123!"
	hash, err := HashPassword(password)
	if err != nil {
		b.Fatal(err)
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		VerifyPassword(password, hash)
	}
}