package crypto

import (
	"crypto/rand"
	"encoding/hex"

	"golang.org/x/crypto/bcrypt"
)

const SALT_BYTES = 16
const BCRYPT_ROUNDS = 12

func GenerateRandomPassword(length int) (string, error) {
	const chars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	for i, v := range b {
		b[i] = chars[v%byte(len(chars))]
	}
	return string(b), nil
}

func GenerateSalt() (string, error) {
	b := make([]byte, SALT_BYTES)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func HashPassword(password, salt string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(salt+password), BCRYPT_ROUNDS)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

func VerifyPassword(password, salt, hash string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(salt+password)) == nil
}

func GenerateSDKKey(prefix string) (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return prefix + hex.EncodeToString(b), nil
}
