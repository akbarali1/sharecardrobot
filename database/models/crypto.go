package models

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"os"
)

// getCardEncryptionKey reads and validates the AES-256 key from CARD_ENCRYPTION_KEY.
func getCardEncryptionKey() ([]byte, error) {
	keyB64 := os.Getenv("CARD_ENCRYPTION_KEY")
	if keyB64 == "" {
		return nil, errors.New("CARD_ENCRYPTION_KEY is required")
	}
	key, err := base64.StdEncoding.DecodeString(keyB64)
	if err != nil || len(key) != 32 {
		return nil, errors.New("CARD_ENCRYPTION_KEY must be a base64-encoded 32-byte value")
	}
	return key, nil
}

func ValidateCardEncryptionConfig() error {
	_, err := getCardEncryptionKey()
	return err
}

// encryptCardNumber encrypts a card number using AES-256-GCM.
// A deterministic nonce derived from HMAC-SHA256(key, plaintext) is used so that
// exact-match SQL lookups remain possible after encryption.
func encryptCardNumber(cardNumber string) string {
	return encryptSecretValue(cardNumber)
}

func encryptExpiryDate(expiryDate string) string {
	return encryptSecretValue(expiryDate)
}

func encryptSecretValue(value string) string {
	key, err := getCardEncryptionKey()
	if err != nil {
		panic(err)
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		panic(err)
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		panic(err)
	}

	// Derive a deterministic nonce: first NonceSize bytes of HMAC-SHA256(key, plaintext).
	mac := hmac.New(sha256.New, key)
	mac.Write([]byte(value))
	nonce := mac.Sum(nil)[:gcm.NonceSize()]

	ciphertext := gcm.Seal(nonce, nonce, []byte(value), nil)
	return base64.StdEncoding.EncodeToString(ciphertext)
}

// decryptCardNumber decrypts a card number that was encrypted with encryptCardNumber.
func decryptCardNumber(value string) string {
	return decryptSecretValue(value)
}

func decryptExpiryDate(value string) string {
	return decryptSecretValue(value)
}

func decryptSecretValue(value string) string {
	key, err := getCardEncryptionKey()
	if err != nil {
		panic(err)
	}

	data, err := base64.StdEncoding.DecodeString(value)
	if err != nil {
		panic(err)
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		panic(err)
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		panic(err)
	}

	nonceSize := gcm.NonceSize()
	if len(data) < nonceSize {
		panic("encrypted value is shorter than GCM nonce size")
	}

	plaintext, err := gcm.Open(nil, data[:nonceSize], data[nonceSize:], nil)
	if err != nil {
		panic(err)
	}
	return string(plaintext)
}
