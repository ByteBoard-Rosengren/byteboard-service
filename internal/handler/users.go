package handler

import (
	"byte-board/internal/middleware"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/rs/zerolog/log"
)

// GET /api/admin/users Handler to get all Users with admin permissions
func (h *Handler) GetAllUsers(w http.ResponseWriter, r *http.Request) {
	log.Info().Msg("GET /users - Getting all users")

	users, err := h.db.GetAllUsers()
	if err != nil {
		log.Error().Err(err).Msg("Failed to get all users")
		writeErrorResponse(w, http.StatusInternalServerError, "Failed to get all users")
		return
	}

	log.Info().Msg("Successfully retrieved all users")
	writeJSONResponse(w, http.StatusOK, users)
}

// GET /api/admin/users/{userId} - Handler to get User by User ID with admin permissions
func (h *Handler) GetUserById(w http.ResponseWriter, r *http.Request) {
	log.Info().Msg("GET /users/{userId} - Getting user by user ID")

	// Get ID
	vars := mux.Vars(r)
	idStr := vars["userId"]

	// Convert int UserID to a string
	id, err := strconv.Atoi(idStr)
	if err != nil {
		log.Warn().Str("ID", idStr).Msg("Invalid user ID format")
		writeErrorResponse(w, http.StatusBadRequest, "Invalid user ID")
		return
	}

	user, err := h.db.GetUserByID(id)
	if err != nil {
		if err.Error() == "user not found" {
			log.Warn().Int("ID", id).Msg("No user with that ID found")
			writeErrorResponse(w, http.StatusNotFound, "User not found")
			return
		}
		log.Error().Err(err).Msg("Failed to get user with that ID")
		writeErrorResponse(w, http.StatusInternalServerError, "Failed to get user")
		return
	}

	log.Info().Int("ID", id).Msg("Successfully retrieved user")
	writeJSONResponse(w, http.StatusOK, user)
}

// GET /api/users/username/{username} - Handler to get User by Username with admin permissions
func (h *Handler) GetUserByUsername(w http.ResponseWriter, r *http.Request) {
	log.Info().Msg("GET /users/username/{username} - Getting user by username")

	// Get username
	vars := mux.Vars(r)
	username := vars["username"]

	user, err := h.db.GetUserByUsername(username)
	if err != nil {
		if err.Error() == "username not found" {
			log.Warn().Str("username", username).Msg("No user with that username found")
			writeErrorResponse(w, http.StatusNotFound, "Username not found")
			return
		}
		log.Error().Err(err).Msg("Failed to get user with that username")
		writeErrorResponse(w, http.StatusInternalServerError, "Failed to get user")
		return
	}

	log.Info().Str("Username", username).Msg("Successfully retrieved user")
	writeJSONResponse(w, http.StatusOK, user)
}

// DELETE /api/users/{userId} - Delete a user and their profile
func (h *Handler) DeleteUser(w http.ResponseWriter, r *http.Request) {
	// Get username from context
	username := middleware.GetUsername(r)
	if username == "" {
		log.Warn().Msg("No username in the context")
		writeErrorResponse(w, http.StatusUnauthorized, "Not authenticated")
		return
	}

	// Get user from the db
	user, err := h.db.GetUserByUsername(username)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get user")
		writeErrorResponse(w, http.StatusInternalServerError, "Failed to get user information")
		return
	}

	// Get the userID string from URL
	vars := mux.Vars(r)
	idStr := vars["userId"]

	// Convert the ID to int
	id, err := strconv.Atoi(idStr)
	if err != nil {
		log.Warn().Str("User ID", idStr).Msg("Invalid User ID format")
		writeErrorResponse(w, http.StatusBadRequest, "Invalid user ID format")
		return
	}

	// Verify user owns the account or is an admin
	if user.ID != id && user.Role != "admin" {
		log.Warn().Msg("User does not own this account")
		writeErrorResponse(w, http.StatusForbidden, "You can only delete your account")
		return
	}

	// Delete the user (cascades to profile, posts, comments)
	if err := h.db.DeleteUser(id); err != nil {
		log.Error().Err(err).Msg("Failed to delete user")
		writeErrorResponse(w, http.StatusInternalServerError, "Failed to delete user")
		return
	}

	// Success
	log.Info().Int("User ID", id).Msg("User account deleted successfully")
	writeJSONResponse(w, http.StatusOK, "User successfully deleted!")
}
