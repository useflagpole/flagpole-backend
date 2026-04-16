package main

import (
	"errors"
	"sync"

	"golang.org/x/crypto/bcrypt"
)

type User struct {
	Email        string
	PasswordHash string
}

type UserStoreType struct {
	mu    sync.RWMutex
	users map[string]User
}

var UserStore = &UserStoreType{
	users: make(map[string]User),
}

func (s *UserStoreType) Register(email, password string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.users[email]; exists {
		return errors.New("email already registered")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	s.users[email] = User{Email: email, PasswordHash: string(hash)}
	return nil
}

func (s *UserStoreType) Authenticate(email, password string) error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	user, exists := s.users[email]
	if !exists {
		// run bcrypt anyway to prevent timing attacks
		bcrypt.CompareHashAndPassword([]byte("$2a$10$invalid"), []byte(password))
		return errors.New("invalid credentials")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return errors.New("invalid credentials")
	}

	return nil
}
