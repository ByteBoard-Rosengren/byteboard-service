package middleware

import (
	"net/http"
	"strings"

	"github.com/rs/zerolog/log"
)

// Holds configuration for CORS middleware
type CORSConfig struct {
	AllowedOrigins []string
}

// CORS adds Cross-Origin Resource Sharing headers to responses with credential support
func CORS(config CORSConfig) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")

			// Validate origin against allowed list
			if isAllowedOrigin(origin, config.AllowedOrigins) {
				// Set specific origin (required for credentials)
				w.Header().Set("Access-Control-Allow-Origin", origin)
				// Enable credentials (cookies, authorization headers)
				w.Header().Set("Access-Control-Allow-Credentials", "true")

				log.Debug().
					Str("origin", origin).
					Str("method", r.Method).
					Str("path", r.URL.Path).
					Msg("CORS: Allowed origin")
			} else if origin != "" {
				// Log rejected origins for security monitoring
				log.Warn().
					Str("origin", origin).
					Str("method", r.Method).
					Str("path", r.URL.Path).
					Msg("CORS: Rejected origin")
			}

			// Set allowed methods
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")

			// Set allowed headers
			w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, Authorization")

			// Security headers
			w.Header().Set("X-Content-Type-Options", "nosniff")
			w.Header().Set("X-Frame-Options", "DENY")
			w.Header().Set("X-XSS-Protection", "1; mode=block")

			// Handle preflight requests
			if r.Method == "OPTIONS" {
				w.WriteHeader(http.StatusOK)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// isAllowedOrigin checks if the origin is in the allowed list
func isAllowedOrigin(origin string, allowed []string) bool {
	if origin == "" {
		return false
	}

	for _, allowedOrigin := range allowed {
		if origin == strings.TrimSpace(allowedOrigin) {
			return true
		}
	}

	return false
}
