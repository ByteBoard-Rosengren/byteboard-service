package handler

import (
	"byte-board/internal/appconfig"
	"byte-board/internal/repository"
	"byte-board/internal/service"
	"encoding/json"
	"net/http"

	"github.com/rs/zerolog/log"
)

type Handler struct {
	db          *repository.DB
	config      *appconfig.Config
	authService *service.AuthService
}

// Create a new instance of a handler
func New(db *repository.DB, cfg *appconfig.Config, authService *service.AuthService) *Handler {
	return &Handler{
		db:          db,
		config:      cfg,
		authService: authService,
	}
}

// Represents an error response
type ErrorResponse struct {
	Error string `json:"error"`
}

// Writes a JSON response
func writeJSONResponse(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	if err := json.NewEncoder(w).Encode(data); err != nil {
		log.Error().Err(err).Msg("Error encoding JSON response")
	}
}

// Writes an error response
func writeErrorResponse(w http.ResponseWriter, status int, message string) {
	log.Warn().Int("status", status).Str("message", message).Msg("Writing error response")
	writeJSONResponse(w, status, ErrorResponse{Error: message})
}
