package cryption

import (
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"fmt"
)

// Structure to parse JSON input
type SecretData struct {
	AccessKey  string  `json:"accessKey"`
	SecretKey  SString `json:"secretKey"`
	Expiration string  `json:"expiration"`
	StatusCode int     `json:"StatusCode"`
}

// SString equivalent in Go
type SString struct {
	CString string `json:"CString"`
	DString string `json:"DString"`
}

// AES decryption function
func (s SString) GetDString() (string, error) {
	// Decode base64 CString
	dummaBytes := []byte{
		'0', '0', '0', '0', '0', '0', '0', '0',
		'0', '0', '0', '0', '0', '0', '0', '0',
	}
	encryptedData, err := base64.StdEncoding.DecodeString(s.CString)
	if err != nil {
		return "", fmt.Errorf("failed to decode base64 string: %v", err)
	}

	// Use the provided DummyChars as key (converted to bytes)
	key := []byte("E8AA3FBB0F512B32") // Must be exactly 16 bytes
	//iv := make([]byte, aes.BlockSize) // IV is 16 bytes of zero (same as .NET)

	// Decrypt AES-128-CBC
	plaintext, err := decryptAES128CBC(encryptedData, key, dummaBytes)
	if err != nil {
		return "", fmt.Errorf("decryption failed: %v", err)
	}

	return string(plaintext), nil
}

// AES-128-CBC decryption with PKCS7 padding removal
func decryptAES128CBC(encryptedData, key, iv []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	if len(encryptedData) < aes.BlockSize {
		return nil, fmt.Errorf("ciphertext too short")
	}

	mode := cipher.NewCBCDecrypter(block, iv)
	decrypted := make([]byte, len(encryptedData))
	mode.CryptBlocks(decrypted, encryptedData)

	// Debug: Print decrypted bytes before removing padding
	//fmt.Println("Decrypted raw bytes:", decrypted)

	// Remove PKCS7 padding
	return removePKCS7Padding(decrypted)
}

// PKCS7 padding removal
func removePKCS7Padding(data []byte) ([]byte, error) {
	if len(data) == 0 {
		return nil, fmt.Errorf("empty data")
	}
	padding := int(data[len(data)-1])
	if padding < 1 || padding > len(data) {
		return nil, fmt.Errorf("invalid padding length")
	}
	return data[:len(data)-padding], nil
}
