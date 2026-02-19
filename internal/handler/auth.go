package handler

import (
	"byte-board/internal/middleware"
	"byte-board/internal/model"
	"encoding/json"
	"net/http"

	"github.com/rs/zerolog/log"
)

// POST /api/register - Register handler
func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	log.Info().Msg("POST /api/register - Registering new user")

	// Parse body request
	var req model.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Warn().Err(err).Msg("Invalid request body")
		writeErrorResponse(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Validate required fields
	if req.Username == "" || req.Password == "" || req.FirstName == "" || req.LastName == "" {
		log.Warn().Msg("Missing required fields")
		writeErrorResponse(w, http.StatusBadRequest, "Username, password, first name, and last name are required")
		return
	}

	// Create user and profile with auth service
	user, profile, token, err := h.authService.Register(req.Username, req.Password, req.FirstName, req.LastName)
	if err != nil {
		// Specific errors
		if err.Error() == "username already exists" {
			log.Warn().Str("username", req.Username).Msg("Username already exists")
			writeErrorResponse(w, http.StatusConflict, "Username already exists")
			return
		}
		if err.Error() == "password must be at least 8 characters long" {
			log.Warn().Msg("Password too short")
			writeErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		log.Error().Err(err).Msg("Failed to register user")
		writeErrorResponse(w, http.StatusInternalServerError, "Failed to register user")
		return
	}

	// Create response
	response := model.AuthResponse{
		Token: token,
		User: model.UserSummary{
			UserID:    user.ID,
			Username:  user.Username,
			Role:      user.Role,
			FirstName: user.FirstName,
			LastName:  user.LastName,
		},
		Profile: profile,
	}

	log.Info().
		Str("username", user.Username).
		Int("user_id", user.ID).
		Msg("User registered successfully")

	writeJSONResponse(w, http.StatusCreated, response)
}

// POST /api/login - Login handler
func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	log.Info().Msg("POST /api/login - User attempting to login")

	// Parse body request
	var req model.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Warn().Err(err).Msg("Invalid request body")
		writeErrorResponse(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Validate required fields
	if req.Username == "" || req.Password == "" {
		log.Warn().Msg("Missing username or password")
		writeErrorResponse(w, http.StatusBadRequest, "Username and password are required")
		return
	}

	// Authenticate user and get JWT token
	token, err := h.authService.Login(req.Username, req.Password)
	if err != nil {
		// Don't reveal whether user or pass was wrong
		log.Warn().Str("username", req.Username).Err(err).Msg("Login failed")
		writeErrorResponse(w, http.StatusUnauthorized, "Invalid username or password")
		return
	}

	// Get user info for response
	user, err := h.db.GetUserByUsername(req.Username)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get user after login")
		writeErrorResponse(w, http.StatusInternalServerError, "Login successful but failed to retrieve user info")
		return
	}

	// Create response
	response := model.AuthResponse{
		Token: token,
		User: model.UserSummary{
			UserID:    user.ID,
			Username:  user.Username,
			Role:      user.Role,
			FirstName: user.FirstName,
			LastName:  user.LastName,
		},
	}

	log.Info().Str("username", user.Username).Int("user_id", user.ID).Msg("User logged in successfully")
	writeJSONResponse(w, http.StatusOK, response)
}

// GET /api/auth/me - GET current user handler
func (h *Handler) GetCurrentUser(w http.ResponseWriter, r *http.Request) {
	log.Info().Msg("GET /api/auth/me - Getting current user")

	// Get username from JWT middleware context
	username := middleware.GetUsername(r)
	if username == "" {
		log.Warn().Msg("No username in context")
		writeErrorResponse(w, http.StatusUnauthorized, "Not authenticated")
		return
	}

	// Get user from database
	user, err := h.db.GetUserByUsername(username)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get current user")
		writeErrorResponse(w, http.StatusInternalServerError, "Failed to get current user")
		return
	}

	// Get user profile from database
	profile, err := h.db.GetProfileByUserId(user.ID)
	if err != nil {
		log.Warn().Err(err).Msg("Failed to get user profile")
		// Continue without profile
	}

	// Create response
	response := map[string]interface{}{
		"user": model.UserSummary{
			UserID:    user.ID,
			Username:  user.Username,
			Role:      user.Role,
			FirstName: user.FirstName,
			LastName:  user.LastName,
		},
		"profile": profile,
	}

	log.Info().Str("username", username).Msg("Successfully retrieved current user")
	writeJSONResponse(w, http.StatusOK, response)
}
