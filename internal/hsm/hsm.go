package hsm

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"

	"github.com/smartedify/auth-service/internal/config"
)

// HSMClient interface for Hardware Security Module operations
type HSMClient interface {
	GenerateKeyPair(keySize int) (*rsa.PrivateKey, error)
	GetPrivateKey(keyID string) (*rsa.PrivateKey, error)
	Sign(keyID string, data []byte) ([]byte, error)
	Encrypt(keyID string, data []byte) ([]byte, error)
	Decrypt(keyID string, data []byte) ([]byte, error)
	RotateKey(keyID string) (*rsa.PrivateKey, error)
}

// MockHSMClient is a mock implementation for development/testing
type MockHSMClient struct {
	config *config.HSMConfig
	keys   map[string]*rsa.PrivateKey
}

func NewMockHSMClient(cfg *config.HSMConfig) HSMClient {
	return &MockHSMClient{
		config: cfg,
		keys:   make(map[string]*rsa.PrivateKey),
	}
}

func (m *MockHSMClient) GenerateKeyPair(keySize int) (*rsa.PrivateKey, error) {
	privateKey, err := rsa.GenerateKey(rand.Reader, keySize)
	if err != nil {
		return nil, fmt.Errorf("failed to generate RSA key pair: %w", err)
	}
	
	// Store the key with the configured key ID
	m.keys[m.config.KeyID] = privateKey
	
	return privateKey, nil
}

func (m *MockHSMClient) GetPrivateKey(keyID string) (*rsa.PrivateKey, error) {
	key, exists := m.keys[keyID]
	if !exists {
		// Generate a new key if it doesn't exist
		return m.GenerateKeyPair(2048)
	}
	return key, nil
}

func (m *MockHSMClient) Sign(keyID string, data []byte) ([]byte, error) {
	privateKey, err := m.GetPrivateKey(keyID)
	if err != nil {
		return nil, err
	}
	
	// This is a simplified signing - in production use proper PKCS#1 v1.5 or PSS
	signature, err := rsa.SignPKCS1v15(rand.Reader, privateKey, 0, data)
	if err != nil {
		return nil, fmt.Errorf("failed to sign data: %w", err)
	}
	
	return signature, nil
}

func (m *MockHSMClient) Encrypt(keyID string, data []byte) ([]byte, error) {
	privateKey, err := m.GetPrivateKey(keyID)
	if err != nil {
		return nil, err
	}
	
	publicKey := &privateKey.PublicKey
	encrypted, err := rsa.EncryptPKCS1v15(rand.Reader, publicKey, data)
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt data: %w", err)
	}
	
	return encrypted, nil
}

func (m *MockHSMClient) Decrypt(keyID string, data []byte) ([]byte, error) {
	privateKey, err := m.GetPrivateKey(keyID)
	if err != nil {
		return nil, err
	}
	
	decrypted, err := rsa.DecryptPKCS1v15(rand.Reader, privateKey, data)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt data: %w", err)
	}
	
	return decrypted, nil
}

func (m *MockHSMClient) RotateKey(keyID string) (*rsa.PrivateKey, error) {
	// Generate new key pair
	newKey, err := m.GenerateKeyPair(2048)
	if err != nil {
		return nil, err
	}
	
	// In a real HSM, you would:
	// 1. Generate new key with new ID
	// 2. Keep old key available for verification
	// 3. Update key rotation metadata
	
	return newKey, nil
}

// AWSCloudHSMClient would be the real implementation for AWS CloudHSM
type AWSCloudHSMClient struct {
	config *config.HSMConfig
	// AWS CloudHSM client would be initialized here
}

func NewAWSCloudHSMClient(cfg *config.HSMConfig) HSMClient {
	// This would initialize the actual AWS CloudHSM client
	// For now, return mock client
	return NewMockHSMClient(cfg)
}

// Utility functions for key management

// ExportPublicKeyPEM exports a public key to PEM format
func ExportPublicKeyPEM(publicKey *rsa.PublicKey) ([]byte, error) {
	pubKeyBytes, err := x509.MarshalPKIXPublicKey(publicKey)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal public key: %w", err)
	}
	
	pubKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: pubKeyBytes,
	})
	
	return pubKeyPEM, nil
}

// ImportPublicKeyPEM imports a public key from PEM format
func ImportPublicKeyPEM(pemData []byte) (*rsa.PublicKey, error) {
	block, _ := pem.Decode(pemData)
	if block == nil {
		return nil, fmt.Errorf("failed to decode PEM block")
	}
	
	pubKey, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse public key: %w", err)
	}
	
	rsaPubKey, ok := pubKey.(*rsa.PublicKey)
	if !ok {
		return nil, fmt.Errorf("not an RSA public key")
	}
	
	return rsaPubKey, nil
}

// KeyManager handles key lifecycle management
type KeyManager struct {
	hsmClient HSMClient
	config    *config.HSMConfig
}

func NewKeyManager(hsmClient HSMClient, cfg *config.HSMConfig) *KeyManager {
	return &KeyManager{
		hsmClient: hsmClient,
		config:    cfg,
	}
}

func (km *KeyManager) GetCurrentPrivateKey() (*rsa.PrivateKey, error) {
	return km.hsmClient.GetPrivateKey(km.config.KeyID)
}

func (km *KeyManager) GetCurrentPublicKey() (*rsa.PublicKey, error) {
	privateKey, err := km.GetCurrentPrivateKey()
	if err != nil {
		return nil, err
	}
	return &privateKey.PublicKey, nil
}

func (km *KeyManager) RotateKeys() error {
	_, err := km.hsmClient.RotateKey(km.config.KeyID)
	return err
}

func (km *KeyManager) SignData(data []byte) ([]byte, error) {
	return km.hsmClient.Sign(km.config.KeyID, data)
}

func (km *KeyManager) EncryptData(data []byte) ([]byte, error) {
	return km.hsmClient.Encrypt(km.config.KeyID, data)
}

func (km *KeyManager) DecryptData(data []byte) ([]byte, error) {
	return km.hsmClient.Decrypt(km.config.KeyID, data)
}