package handler

import (
	"net/http"

	"github.com/rs/zerolog/log"
)

// GET /api/health - Check the health of the service
func (h *Handler) GetHealth(w http.ResponseWriter, r *http.Request) {

	log.Info().Msg("GET /api/health - Getting Service Health")
	writeJSONResponse(w, http.StatusOK, map[string]string{"status": "ok"})
}
