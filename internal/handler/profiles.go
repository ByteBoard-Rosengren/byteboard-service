package handler

import (
	"byte-board/internal/middleware"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/rs/zerolog/log"
)

// GET /api/profiles - Handler to get all profiles
func (h *Handler) GetAllProfiles(w http.ResponseWriter, r *http.Request) {
	log.Info().Msg("GET /profiles - Getting all profiles")

	profiles, err := h.db.GetAllProfiles()
	if err != nil {
		log.Error().Err(err).Msg("Failed to get all profiles")
		writeErrorResponse(w, http.StatusInternalServerError, "Failed to get profiles")
		return
	}

	log.Info().Int("Count", len(profiles)).Msg("Successfully retrieved all profiles")
	writeJSONResponse(w, http.StatusOK, profiles)
}

// GET /api/profiles/{userId} - Handler to get profile by User ID
func (h *Handler) GetProfileByUserId(w http.ResponseWriter, r *http.Request) {
	log.Info().Msg("GET /profiles/{userId} - Getting profile by user ID")

	// Get userID
	vars := mux.Vars(r)
	idStr := vars["userId"]

	// Convert string user ID to an int
	id, err := strconv.Atoi(idStr)
	if err != nil {
		log.Warn().Str("User ID", idStr).Msg("Invalid user ID format")
		writeErrorResponse(w, http.StatusBadRequest, "Invalid ID format")
		return
	}

	profile, err := h.db.GetProfileByUserId(id)
	if err != nil {
		if err.Error() == "profile not found" {
			log.Warn().Int("ID", id).Msg("Profile not found")
			writeErrorResponse(w, http.StatusNotFound, "Profile not found")
			return
		}
		log.Error().Err(err).Msg("Error getting profile")
		writeErrorResponse(w, http.StatusInternalServerError, "Failed to get profile")
		return
	}

	log.Info().Int("ID", id).Msg("Successfully retrieved profile")
	writeJSONResponse(w, http.StatusOK, profile)
}

// PUT /api/profiles/{userId} - Handler to update profile
func (h *Handler) UpdateProfile(w http.ResponseWriter, r *http.Request) {
	log.Info().Msg("PUT /api/profiles/{userId} - Updating profile")

	// Get authenticated username from context
	username := middleware.GetUsername(r)
	if username == "" {
		log.Warn().Msg("No username in the context")
		writeErrorResponse(w, http.StatusUnauthorized, "Not authenticated")
		return
	}

	// Get the user from the db
	user, err := h.db.GetUserByUsername(username)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get user")
		writeErrorResponse(w, http.StatusInternalServerError, "Failed to get user")
		return
	}

	// Get UserID from req URL
	vars := mux.Vars(r)
	idStr := vars["userId"]

	// Convert string ID to int
	id, err := strconv.Atoi(idStr)
	if err != nil {
		log.Warn().Str("User ID", idStr).Msg("Invalid user ID format in URL")
		writeErrorResponse(w, http.StatusBadRequest, "Invalid ID format")
		return
	}

	// Get existing profile from the db
	existingProfile, err := h.db.GetProfileByUserId(id)
	if err != nil {
		if err.Error() == "profile not found" {
			log.Warn().Int("User ID", id).Msg("profile not found")
			writeErrorResponse(w, http.StatusNotFound, "Profile not found")
			return
		}
		log.Error().Err(err).Msg("failed to get profile")
		writeErrorResponse(w, http.StatusInternalServerError, "Failed to get profile")
		return
	}

	// Verify the user owns the profile
	if user.ID != existingProfile.UserId {
		log.Warn().Int("Profile ID", existingProfile.UserId).Int("User ID", user.ID).Msg("User does not own this profile")
		writeErrorResponse(w, http.StatusForbidden, "You can only update your profile")
		return
	}

	// Parse request body
	var req struct {
		FirstName  string `json:"first_name"`
		LastName   string `json:"last_name"`
		Email      string `json:"email"`
		GithubLink string `json:"github_link"`
		City       string `json:"city"`
		State      string `json:"state"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Warn().Msg("Missing required field")
		writeErrorResponse(w, http.StatusBadRequest, "Missing at least one of the required fields, Firstname, Lastname, Email, Github Link, City, or State")
		return
	}

	// Update profile object with new data
	existingProfile.FirstName = req.FirstName
	existingProfile.LastName = req.LastName
	existingProfile.Email = req.Email
	existingProfile.GithubLink = req.GithubLink
	existingProfile.City = req.City
	existingProfile.State = req.State

	// Call the database to update the profile
	if err := h.db.UpdateProfile(existingProfile); err != nil {
		log.Error().Err(err).Msg("Failed to update profile")
		writeErrorResponse(w, http.StatusInternalServerError, "Failed to update profile")
		return
	}

	// Success
	log.Info().Int("User ID", id).Msg("Successfully updated profile")
	writeJSONResponse(w, http.StatusOK, existingProfile)
}
