package auth

import (
	"byte-board/internal/model"
	"errors"
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

// BCrypt cost factor - determines hash complexity
// Higher values = more secure but slower
const DefaultCost = 10

// Generates a BCRYPT hash from a plaintext password
func HashPassword(password string) (string, error) {
	// Validate password
	if password == "" {
		return "", model.ErrPasswordEmpty
	}

	if len(password) > 72 {
		return "", model.ErrPasswordTooLong
	}

	// Generate hash using DefaultCost
	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(password), DefaultCost)
	if err != nil {
		return "", fmt.Errorf("failed to hash password: %w", err)
	}

	return string(hashedBytes), nil
}

// Compare a plaintext password with a bcrypt hash
func CheckPassword(password, hashedPassword string) bool {
	// Compare password with hash. Returns nil on success and error on failure
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))

	// Return false if there is an error
	return err == nil
}

// Like CheckPassword but returns the error
// Useful if you need to distinguish between wrong password vs other error
func CheckPasswordWithError(password, hashedPassword string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
}

// Validate password meets minimum requirements
func ValidatePasswordStrength(password string) error {
	// Validate password is not too long or too short
	if len(password) < 6 {
		return errors.New("password must be at least 6 characters long")
	}
	if len(password) > 72 {
		return model.ErrPasswordTooLong
	}

	return nil
}
