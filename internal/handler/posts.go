package handler

import (
	"byte-board/internal/middleware"
	"byte-board/internal/model"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"github.com/rs/zerolog/log"
)

// GET /api/posts - Handler to get all posts
func (h *Handler) GetAllPosts(w http.ResponseWriter, r *http.Request) {
	log.Info().Msg("GET /posts - Getting all posts")

	posts, err := h.db.GetAllPosts()
	if err != nil {
		log.Error().Err(err).Msg("Error getting all posts")
		writeErrorResponse(w, http.StatusInternalServerError, "Failed to get all posts")
		return
	}

	log.Info().Int("count", len(posts)).Msg("Successfully retrieved all posts")
	writeJSONResponse(w, http.StatusOK, posts)
}

// GET /api/posts/{postId} - Handler to get post by ID
func (h *Handler) GetPostById(w http.ResponseWriter, r *http.Request) {
	log.Info().Msg("GET /posts/{postId} - Getting a post by post ID")

	vars := mux.Vars(r)
	idStr := vars["postId"]

	// Convert the ID from string to an int
	id, err := strconv.Atoi(idStr)
	if err != nil {
		log.Warn().Str("ID", idStr).Msg("Invalid post ID format")
		writeErrorResponse(w, http.StatusBadRequest, "Invalid post ID")
		return
	}

	post, err := h.db.GetPostById(id)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get post by ID")
		writeErrorResponse(w, http.StatusInternalServerError, "Failed to get post by ID")
		return
	}

	viewerId := 0

	// Get authenticated user from context
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
		writeErrorResponse(w, http.StatusInternalServerError, "failed to get user")
		return
	}

	viewerId = user.ID

	counts, err := h.db.GetPostReactionCounts(id, viewerId)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get post reaction counts")
		writeErrorResponse(w, http.StatusInternalServerError, "Failed to get reaction counts")
		return
	}

	log.Info().Int("Post ID", id).Msg("Successfully retrieved post by ID")
	writeJSONResponse(w, http.StatusOK, model.PostResponse{Post: *post, Reactions: counts})
}

// GET /api/posts/user/{userId} - Handler to get all posts by UserID
func (h *Handler) GetPostsByUserId(w http.ResponseWriter, r *http.Request) {
	log.Info().Msg("GET /posts/user/{userId} - Getting all posts by user ID")

	vars := mux.Vars(r)
	idStr := vars["userId"]

	// Convert string ID into an int
	id, err := strconv.Atoi(idStr)
	if err != nil {
		log.Warn().Str("ID", idStr).Msg("Invalid user ID format")
		writeErrorResponse(w, http.StatusBadRequest, "Invalid user ID")
		return
	}

	posts, err := h.db.GetPostsByUserId(id)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get posts from that user")
		writeErrorResponse(w, http.StatusInternalServerError, "Failure to get posts with that user ID")
		return
	}

	log.Info().Int("Count", len(posts)).Msg("Successfully retrieved posts from user ID")
	writeJSONResponse(w, http.StatusOK, posts)
}

// POST /api/posts - Create new post
func (h *Handler) CreatePost(w http.ResponseWriter, r *http.Request) {
	log.Info().Msg("POST /api/posts - Creating new post")

	// Get authenticated user from JWT mware context
	username := middleware.GetUsername(r)
	if username == "" {
		log.Warn().Msg("No username in context")
		writeErrorResponse(w, http.StatusUnauthorized, "Not authenticated")
		return
	}

	// Get user from db
	user, err := h.db.GetUserByUsername(username)
	if err != nil {
		log.Error().Err(err).Msg("failed to get user")
		writeErrorResponse(w, http.StatusInternalServerError, "failed to get user info")
		return
	}

	// Parse body request
	var req struct {
		Title   string `json:"title"`
		Content string `json:"content"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Warn().Err(err).Msg("Invalid request body")
		writeErrorResponse(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Validate input
	if req.Title == "" || req.Content == "" {
		log.Warn().Msg("Missing required fields")
		writeErrorResponse(w, http.StatusBadRequest, "Title and content are required")
		return
	}

	// Create post object
	post := &model.Post{
		UserId:     user.ID,
		Title:      req.Title,
		Content:    req.Content,
		Author:     user.Username,
		DatePosted: time.Now(),
	}

	// Call db to create post
	if err := h.db.CreatePost(post); err != nil {
		log.Error().Err(err).Msg("failed to create post")
		writeErrorResponse(w, http.StatusInternalServerError, "Failed to create post")
		return
	}

	log.Info().Str("title", post.Title).Msg("Post created successfully")
	writeJSONResponse(w, http.StatusCreated, post)
}

// PUT /api/posts/{postId} - Update post
func (h *Handler) UpdatePost(w http.ResponseWriter, r *http.Request) {
	log.Info().Msg("PUT /api/posts/{postId} - Updating a post")

	// Get authenticated user from context
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
		writeErrorResponse(w, http.StatusInternalServerError, "failed to get user")
		return
	}

	// Get post ID from URL params
	vars := mux.Vars(r)
	idStr := vars["postId"]

	// Convert string ID into int
	id, err := strconv.Atoi(idStr)
	if err != nil {
		log.Warn().Str("post_id", idStr).Msg("Invalid post ID format")
		writeErrorResponse(w, http.StatusBadRequest, "Invalid ID format")
		return
	}

	// Get existing post from the db
	existingPost, err := h.db.GetPostById(id)
	if err != nil {
		if err.Error() == "post not found" {
			log.Warn().Int("postId", id).Msg("post not found")
			writeErrorResponse(w, http.StatusNotFound, "Post not found")
			return
		}
		log.Error().Err(err).Msg("failed to get post")
		writeErrorResponse(w, http.StatusInternalServerError, "Failed to get post")
		return
	}

	// Verify the user owns the post (holy cow... long function)
	if existingPost.UserId != user.ID {
		log.Warn().Int("userId", user.ID).Int("postId", existingPost.PostId).Msg("User does not own this post")
		writeErrorResponse(w, http.StatusForbidden, "You can only update your own posts")
		return
	}

	// Parse request body
	var req struct {
		Title   string `json:"title"`
		Content string `json:"content"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Warn().Err(err).Msg("Invalid request body")
		writeErrorResponse(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Validate input
	if req.Title == "" || req.Content == "" {
		log.Warn().Msg("Missing required fields")
		writeErrorResponse(w, http.StatusBadRequest, "Title and content are required")
		return
	}

	// Update post object with new data
	existingPost.Title = req.Title
	existingPost.Content = req.Content

	// Call database to update post
	if err := h.db.UpdatePost(existingPost); err != nil {
		log.Error().Err(err).Msg("failed to update post")
		writeErrorResponse(w, http.StatusInternalServerError, "Failed to update post")
		return
	}

	// Success
	log.Info().Int("postId", id).Str("title", existingPost.Title).Msg("Post updated successfully")
	writeJSONResponse(w, http.StatusOK, existingPost)
}

// DELETE /api/posts/{postId} - Handler to delete a post
func (h *Handler) DeletePost(w http.ResponseWriter, r *http.Request) {
	log.Info().Msg("DELETE /api/posts/{postId} - Deleting post")

	// Get authenticated user from context
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
		writeErrorResponse(w, http.StatusInternalServerError, "Failed to get user")
		return
	}

	// Get the string post ID
	vars := mux.Vars(r)
	idStr := vars["postId"]

	// Conver string postID to an int
	id, err := strconv.Atoi(idStr)
	if err != nil {
		log.Warn().Str("PostID", idStr).Msg("Invalid post ID format")
		writeErrorResponse(w, http.StatusBadRequest, "Invalid ID format")
		return
	}

	// Get existing post from the db
	existingPost, err := h.db.GetPostById(id)
	if err != nil {
		if err.Error() == "post not found" {
			log.Warn().Int("PostID", id).Msg("post not found")
			writeErrorResponse(w, http.StatusNotFound, "Post not found")
			return
		}
		log.Error().Err(err).Msg("failed to get post")
		writeErrorResponse(w, http.StatusInternalServerError, "Failed to get post")
		return
	}

	// Verify the user owns the post or user deleting post is admin
	if existingPost.UserId != user.ID && user.Role != "admin" {
		log.Warn().Int("PostID", id).Int("UserID", user.ID).Msg("User does not own this post")
		writeErrorResponse(w, http.StatusForbidden, "You can only delete your own posts")
		return
	}

	// Call the database to delete the post
	if err := h.db.DeletePost(id); err != nil {
		log.Error().Err(err).Msg("failed to delete post")
		writeErrorResponse(w, http.StatusInternalServerError, "Failed to delete post")
		return
	}

	log.Info().Int("PostID", id).Msg("Post deleted successfully")
	writeJSONResponse(w, http.StatusOK, map[string]string{"message": "Post deleted successfully"})
}
