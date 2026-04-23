package models

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"log"
	"os"
)

// getCardEncryptionKey reads and validates the AES-256 key from CARD_ENCRYPTION_KEY.
// Returns nil key (no error) when the env var is unset, enabling plaintext fallback.
func getCardEncryptionKey() ([]byte, error) {
	keyB64 := os.Getenv("CARD_ENCRYPTION_KEY")
	if keyB64 == "" {
		return nil, nil
	}
	key, err := base64.StdEncoding.DecodeString(keyB64)
	if err != nil || len(key) != 32 {
		return nil, errors.New("CARD_ENCRYPTION_KEY must be a base64-encoded 32-byte value")
	}
	return key, nil
}

// encryptCardNumber encrypts a card number using AES-256-GCM.
// A deterministic nonce derived from HMAC-SHA256(key, plaintext) is used so that
// exact-match SQL lookups remain possible after encryption.
// Returns the input unchanged when CARD_ENCRYPTION_KEY is not set.
func encryptCardNumber(cardNumber string) string {
	key, err := getCardEncryptionKey()
	if err != nil {
		log.Printf("card encryption key error: %v", err)
		return cardNumber
	}
	if key == nil {
		return cardNumber
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		log.Printf("card encryption cipher error: %v", err)
		return cardNumber
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		log.Printf("card encryption gcm error: %v", err)
		return cardNumber
	}

	// Derive a deterministic nonce: first NonceSize bytes of HMAC-SHA256(key, plaintext).
	mac := hmac.New(sha256.New, key)
	mac.Write([]byte(cardNumber))
	nonce := mac.Sum(nil)[:gcm.NonceSize()]

	ciphertext := gcm.Seal(nonce, nonce, []byte(cardNumber), nil)
	return base64.StdEncoding.EncodeToString(ciphertext)
}

// decryptCardNumber decrypts a card number that was encrypted with encryptCardNumber.
// Returns the input unchanged when CARD_ENCRYPTION_KEY is not set or when decryption
// fails (to allow a graceful migration from pre-existing plaintext values).
func decryptCardNumber(value string) string {
	key, err := getCardEncryptionKey()
	if err != nil {
		log.Printf("card encryption key error: %v", err)
		return value
	}
	if key == nil {
		return value
	}

	data, err := base64.StdEncoding.DecodeString(value)
	if err != nil {
		// Not base64 – likely a pre-existing plaintext value.
		return value
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		log.Printf("card decryption cipher error: %v", err)
		return value
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		log.Printf("card decryption gcm error: %v", err)
		return value
	}

	nonceSize := gcm.NonceSize()
	if len(data) < nonceSize {
		return value
	}

	plaintext, err := gcm.Open(nil, data[:nonceSize], data[nonceSize:], nil)
	if err != nil {
		// Decryption failed – return as-is (plaintext migration period).
		return value
	}
	return string(plaintext)
}
