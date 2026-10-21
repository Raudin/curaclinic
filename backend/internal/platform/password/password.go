// Package password hashes and verifies user passwords.
package password

import (
	"errors"
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

// MaxLength is bcrypt's maximum accepted password length.
const MaxLength = 72

// Hash returns a bcrypt hash using the supplied cost.
func Hash(plain string, cost int) (string, error) {
	if err := Validate(plain); err != nil {
		return "", err
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(plain), cost)
	if err != nil {
		return "", fmt.Errorf("hash password: %w", err)
	}
	return string(hash), nil
}

// Verify reports whether the plain password matches the hash.
func Verify(hash, plain string) bool {
	if len(plain) == 0 || len(plain) > MaxLength {
		return false
	}
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(plain)) == nil
}

// Validate enforces the password length policy.
func Validate(plain string) error {
	if len(plain) < 12 {
		return errors.New("password must be at least 12 characters")
	}
	if len(plain) > MaxLength {
		return errors.New("password must be at most 72 characters")
	}
	return nil
}
