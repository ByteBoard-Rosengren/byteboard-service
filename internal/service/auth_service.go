package service

import (
	"byte-board/internal/auth"
	"byte-board/internal/model"
	"byte-board/internal/repository"
	"fmt"
	"time"
)

// Handles authentication business logic
type AuthService struct {
	db            *repository.DB
	tokenProvider *auth.TokenProvider
}

// Creates new authentication service
func NewAuthService(db *repository.DB, tokenProvider *auth.TokenProvider) *AuthService {
	return &AuthService{
		db:            db,
		tokenProvider: tokenProvider,
	}
}

// Login - Authenticate user and return JWT token
func (s *AuthService) Login(username, password string) (string, error) {
	// Get user from database
	user, err := s.db.GetUserByUsername(username)
	if err != nil {
		return "", fmt.Errorf("invalid credentials")
	}

	// Verify password
	if !auth.CheckPassword(password, user.HashedPassword) {
		return "", fmt.Errorf("invalid credentials")
	}

	// Generate JWT token
	token, err := s.tokenProvider.CreateToken(user.Username, user.Role)
	if err != nil {
		return "", fmt.Errorf("failed to generate token: %w", err)
	}

	return token, nil
}

// Creates new account
func (s *AuthService) Register(username, password, firstName, lastName string) (*model.User, *model.Profile, string, error) {
	// Validate password strength
	if err := auth.ValidatePasswordStrength(password); err != nil {
		return nil, nil, "", fmt.Errorf("invalid password: %w", err)
	}

	// Check if username already exists
	exists, err := s.db.UserExists(username)
	if err != nil {
		return nil, nil, "", fmt.Errorf("failed to check username availability: %w", err)
	}
	if exists {
		return nil, nil, "", fmt.Errorf("username already exists")
	}

	// Hash password
	hashedPassword, err := auth.HashPassword(password)
	if err != nil {
		return nil, nil, "", fmt.Errorf("failed to hash password: %w", err)
	}

	// Create user object
	user := &model.User{
		Username:       username,
		HashedPassword: hashedPassword,
		Role:           "user",
		FirstName:      firstName,
		LastName:       lastName,
	}

	// Save to database
	if err := s.db.CreateUser(user); err != nil {
		return nil, nil, "", fmt.Errorf("failed to create user: %w", err)
	}

	// Create profile for user
	profile := &model.Profile{
		UserId:         user.ID,
		FirstName:      firstName,
		LastName:       lastName,
		Email:          "",
		GithubLink:     "",
		City:           "",
		State:          "",
		DateRegistered: time.Now(),
	}

	// Add new profile to the database
	createdProfile, err := s.db.CreateProfile(profile)
	if err != nil {
		return nil, nil, "", fmt.Errorf("failed to create profile: %w", err)
	}

	// Generate JWT token so the user is immediately authenticated after registration
	token, err := s.tokenProvider.CreateToken(user.Username, user.Role)
	if err != nil {
		return nil, nil, "", fmt.Errorf("failed to generate token: %w", err)
	}

	// user.ID is now populated by CreateUser bc of RETURNING clause
	return user, createdProfile, token, nil
}

// Change a user's password
func (s *AuthService) ChangePassword(userId int, oldPass, newPass string) error {
	// Get user
	user, err := s.db.GetUserByID(userId)
	if err != nil {
		return fmt.Errorf("failed to get user: %w", err)
	}

	// Verify old password
	if !auth.CheckPassword(oldPass, user.HashedPassword) {
		return fmt.Errorf("invalid current password")
	}

	// Validate new password
	if err := auth.ValidatePasswordStrength(newPass); err != nil {
		return err
	}

	// Hash new password
	hashedPass, err := auth.HashPassword(newPass)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	// Update user
	user.HashedPassword = hashedPass
	if err := s.db.UpdateUser(user); err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}

	return nil
}

// Checks if JWT token is valid
func (s *AuthService) ValidateToken(tokenString string) error {
	return s.tokenProvider.ValidateToken(tokenString)
}

// Extracts user information from a JWT token
func (s *AuthService) GetUserFromToken(tokenString string) (*model.User, error) {
	// Parse token
	claims, err := s.tokenProvider.ParseToken(tokenString)
	if err != nil {
		return nil, err
	}

	// Get user from database
	user, err := s.db.GetUserByUsername(claims.Username)
	if err != nil {
		return nil, fmt.Errorf("user not found")
	}

	return user, nil
}
