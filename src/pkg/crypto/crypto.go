package crypto

import (
	"crypto/rand"
	"encoding/hex"

	"golang.org/x/crypto/bcrypt"
)

const saltBytes = 16
const bcryptRounds = 12

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
	b := make([]byte, saltBytes)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func HashPassword(password, salt string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(salt+password), bcryptRounds)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

func VerifyPassword(password, salt, hash string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(salt+password)) == nil
}
