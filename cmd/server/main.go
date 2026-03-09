package main

import (
	"byte-board/internal/appconfig"
	"byte-board/internal/auth"
	"byte-board/internal/handler"
	"byte-board/internal/middleware"
	"byte-board/internal/service"
	"net/http"
	"os"
	"time"

	database "byte-board/internal/repository"

	"github.com/gorilla/mux"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func main() {
	// Setup Zerologger
	zerolog.TimeFieldFormat = time.RFC3339
	log.Logger = zerolog.New(zerolog.ConsoleWriter{
		Out:        os.Stderr,
		TimeFormat: "2006-01-02 15:04:05",
		FormatTimestamp: func(i interface{}) string {
			t, _ := time.Parse(time.RFC3339, i.(string))
			return t.Local().Format("2006-01-02 15:04:05")
		},
	}).
		With().
		Timestamp().
		Logger()

	log.Info().Msg("Starting Byte Board Backend Service!")

	// Load configurations
	cfg, err := appconfig.Load()
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to load configuration")
	}

	// Initialize database
	db, err := database.New(cfg)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to initialize database")
	}
	defer db.Close()

	// Initialize JWT token provider
	jwtConfig := auth.JWTConfig{
		SecretKey:       cfg.JWTSecret,
		ExpirationHours: cfg.JWTExpirationHours,
	}
	tokenProvider := auth.NewTokenProvider(jwtConfig)
	log.Info().Msg("JWT token provider initialized")

	// Initialize auth service
	authService := service.NewAuthService(db, tokenProvider)
	log.Info().Msg("Auth service initialized")

	// Initialize auth middleware
	authMiddleware := middleware.NewAuthMiddleware(tokenProvider)
	log.Info().Msg("Auth middleware initialized")

	// Initialize handlers with auth service
	handler := handler.New(db, cfg, authService)

	// Set up router with middlewear
	router := setupRouter(handler, authMiddleware)

	// Initialize CORS middleware with configuration
	corsConfig := middleware.CORSConfig{
		AllowedOrigins: cfg.GetAllowedOrigins(),
	}

	// Apply middleware chain: Recover -> Logging -> CORS -> Router
	httpHandler := middleware.Recovery(
		middleware.Logging(
			middleware.CORS(corsConfig)(router),
		),
	)

	// Start server
	log.Info().Str("port", cfg.Port).Msg("Byte Board Service starting")

	server := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      httpHandler,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	log.Fatal().Err(server.ListenAndServe()).Msg("Server failed to start")
}

// Setup router configures all of the API routes
func setupRouter(h *handler.Handler, authMiddleware *middleware.AuthMiddleware) *mux.Router {
	router := mux.NewRouter()

	// Set up API routes
	api := router.PathPrefix("/api").Subrouter()

	// Set up protected routes (JWT Required)
	protected := api.PathPrefix("").Subrouter()
	protected.Use(authMiddleware.JWTAuth)

	// Set up admin routes
	admin := api.PathPrefix("/admin").Subrouter()
	admin.Use(authMiddleware.JWTAuth)
	admin.Use(middleware.RequireRole("admin"))

	// Login/Register endpoints
	api.HandleFunc("/register", h.Register).Methods("POST")
	api.HandleFunc("/login", h.Login).Methods("POST")

	// Comment endpoints
	// GET
	protected.HandleFunc("/comments", h.GetAllComments).Methods("GET")
	protected.HandleFunc("/posts/{postId}/comments", h.GetCommentsOnPost).Methods("GET")
	protected.HandleFunc("/comments/{commentId}", h.GetCommentById).Methods("GET")
	// POST
	protected.HandleFunc("/posts/{postId}/comments", h.CreateComment).Methods("POST")
	// PUT
	protected.HandleFunc("/comments/{commentId}", h.UpdateComment).Methods("PUT")
	// DELETE
	protected.HandleFunc("/comments/{commentId}", h.DeleteComment).Methods("DELETE")

	// Post endpoints
	// GET
	protected.HandleFunc("/posts", h.GetAllPosts).Methods("GET")
	protected.HandleFunc("/posts/{postId}", h.GetPostById).Methods("GET")
	protected.HandleFunc("/posts/user/{userId}", h.GetPostsByUserId).Methods("GET")
	// POST
	protected.HandleFunc("/posts", h.CreatePost).Methods("POST")
	// PUT
	protected.HandleFunc("/posts/{postId}", h.UpdatePost).Methods("PUT")
	// DELETE
	protected.HandleFunc("/posts/{postId}", h.DeletePost).Methods("DELETE")

	// Profile endpoints
	protected.HandleFunc("/profiles", h.GetAllProfiles).Methods("GET")
	protected.HandleFunc("/profiles/{userId}", h.GetProfileByUserId).Methods("GET")
	// PUT
	protected.HandleFunc("/profiles/{userId}", h.UpdateProfile).Methods("PUT")

	// User endpoints
	protected.HandleFunc("/auth/me", h.GetCurrentUser).Methods("GET")
	// DELETE
	protected.HandleFunc("/users/{userId}", h.DeleteUser).Methods("DELETE")

	// Notification endpoints
	protected.HandleFunc("/notifications", h.GetNotificationsByUserId).Methods("GET")
	protected.HandleFunc("/notifications/{notificationId}/read", h.MarkNotificationAsRead).Methods("PUT")

	// Comment and post reactions
	protected.HandleFunc("/posts/{postId}/react", h.ReactToPost).Methods("POST")
	protected.HandleFunc("/comments/{commentId}/react", h.ReactToComment).Methods("POST")

	// User management (Admin only)
	admin.HandleFunc("/users", h.GetAllUsers).Methods("GET")
	admin.HandleFunc("/users/{userId}", h.GetUserById).Methods("GET")
	admin.HandleFunc("/users/username/{username}", h.GetUserByUsername).Methods("GET")

	return router
}
