package handler

import (
	"byte-board/internal/appconfig"
	"byte-board/internal/middleware"
	"byte-board/internal/model"
	"byte-board/internal/repository"
	"byte-board/internal/service"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
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

// #region Comment handlers

// GET /api/comments - Handler to get all comments
func (h *Handler) GetAllComments(w http.ResponseWriter, r *http.Request) {
	log.Info().Msg("GET /comments - Getting all comments")

	comments, err := h.db.GetAllComments()
	if err != nil {
		log.Error().Err(err).Msg("Error getting comments")
		writeErrorResponse(w, http.StatusInternalServerError, "failed to get comments")
		return
	}

	log.Info().Int("count", len(comments)).Msg("Successfully retrieved comments!")
	writeJSONResponse(w, http.StatusOK, comments)
}

// GET /api/comments/{commentId} - Handler to get a comment by comment ID
func (h *Handler) GetCommentById(w http.ResponseWriter, r *http.Request) {
	log.Info().Msg("GET /comments/{CommentID} - Getting comment by its ID")

	vars := mux.Vars(r)
	idStr := vars["commentId"]

	log.Info().Str("comment_id", idStr).Msg("GET /comments/{CommentID} - Getting comment by ID")

	// Convert id string into an int
	id, err := strconv.Atoi(idStr)
	if err != nil {
		log.Warn().Str("id", idStr).Msg("Invalid ID format")
		writeErrorResponse(w, http.StatusBadRequest, "Invalid ID format")
		return
	}

	// Get comment by id from the database
	comment, err := h.db.GetCommentById(id)
	if err != nil {
		if err.Error() == "comment not found" {
			log.Warn().Int("ID", id).Msg("Comment with that ID not found")
			writeErrorResponse(w, http.StatusNotFound, "Comment not found")
			return
		}
		log.Error().Err(err).Msg("Failed to get comment by ID")
		writeErrorResponse(w, http.StatusInternalServerError, "Failed to get that comment")
		return
	}

	log.Info().Int("ID", id).Msg("Successfully retrieved the comment")
	writeJSONResponse(w, http.StatusOK, comment)
}

// GET /api/post/{postId}/comments - Handler to get all of the comments on a post
func (h *Handler) GetCommentsOnPost(w http.ResponseWriter, r *http.Request) {
	log.Info().Msg("GET /post/{postId}/comments - Getting comments on post")

	vars := mux.Vars(r)
	idStr := vars["postId"]

	// Convert the ID string into an int
	id, err := strconv.Atoi(idStr)
	if err != nil {
		log.Warn().Str("id", idStr).Msg("Invalid post ID format")
		writeErrorResponse(w, http.StatusBadRequest, "Invalid Post ID")
		return
	}

	comments, err := h.db.GetCommentsByPost(id)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get all comments on the post")
		writeErrorResponse(w, http.StatusInternalServerError, "failed to get comments on post")
		return
	}

	log.Info().Int("count", len(comments)).Msg("Successfully retrieved comments on post")
	writeJSONResponse(w, http.StatusOK, comments)

}

// POST /api/post/{postId}/comments - Creating comment on a post
func (h *Handler) CreateComment(w http.ResponseWriter, r *http.Request) {
	log.Info().Msg("Creating comment on a post")

	// Get the post ID as a string from URL params
	vars := mux.Vars(r)
	postIdStr := vars["postId"]

	// Convert post ID string into int
	postId, err := strconv.Atoi(postIdStr)
	if err != nil {
		log.Warn().Str("Post ID", postIdStr).Msg("Invalid Post ID format")
		writeErrorResponse(w, http.StatusBadRequest, "Invalid post ID")
		return
	}

	// Get username
	username := middleware.GetUsername(r)
	if username == "" {
		log.Warn().Msg("No username in that context")
		writeErrorResponse(w, http.StatusUnauthorized, "Unauthorized user")
		return
	}

	// Get user from db
	user, err := h.db.GetUserByUsername(username)
	if err != nil {
		log.Error().Err(err).Msg("failed to get user")
		writeErrorResponse(w, http.StatusInternalServerError, "Failed to get user info")
		return
	}

	// Verify post exists
	_, err = h.db.GetPostById(postId)
	if err != nil {
		if err.Error() == "post not found" {
			log.Warn().Int("Post ID", postId).Msg("Post not found")
			writeErrorResponse(w, http.StatusNotFound, "Post not found")
			return
		}
		log.Error().Err(err).Msg("Failed to verify post")
		writeErrorResponse(w, http.StatusInternalServerError, "Failed to verify post existence")
		return
	}

	// Parse the request body
	var req struct {
		Content string `json:"content"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Warn().Err(err).Msg("Invalid request body")
		writeErrorResponse(w, http.StatusBadRequest, "Invalid req body")
		return
	}

	// Validate input
	if req.Content == "" {
		log.Warn().Msg("Missing required content field")
		writeErrorResponse(w, http.StatusBadRequest, "Content is required")
		return
	}

	// Create comment object
	comment := model.Comment{
		UserId:     user.ID,
		PostId:     postId,
		Content:    req.Content,
		Author:     user.Username,
		DatePosted: time.Now(),
	}

	// Call database to create comment
	if err := h.db.CreateComment(&comment, postId); err != nil {
		log.Error().Err(err).Msg("Failed to create comment")
		writeErrorResponse(w, http.StatusInternalServerError, "Failed to create comment")
		return
	}

	// Success
	log.Info().Int("Comment ID", comment.CommentId).Msg("Successfully added comment to post")
	writeJSONResponse(w, http.StatusCreated, comment)

	// Go routine to create the notification
	go h.createNotification(comment)
}

// PUT /api/comments/{commentId} - Update comment
func (h *Handler) UpdateComment(w http.ResponseWriter, r *http.Request) {
	log.Info().Msg("PUT /api/comments/{commentId} - Updating comment")

	// Verify authenticated user
	username := middleware.GetUsername(r)
	if username == "" {
		log.Warn().Msg("No username in context")
		writeErrorResponse(w, http.StatusUnauthorized, "Not authenticated")
		return
	}

	// Get user from db
	user, err := h.db.GetUserByUsername(username)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get user info")
		writeErrorResponse(w, http.StatusInternalServerError, "Failed to get user info")
		return
	}

	// Get comment ID string from URL
	vars := mux.Vars(r)
	idStr := vars["commentId"]

	// Convert comment ID string to int
	id, err := strconv.Atoi(idStr)
	if err != nil {
		log.Warn().Str("Comment ID", idStr).Msg("Invalid Comment ID format")
		writeErrorResponse(w, http.StatusBadRequest, "Invalid Comment ID")
		return
	}

	// Get existing comment from db
	existingComment, err := h.db.GetCommentById(id)
	if err != nil {
		if err.Error() == "comment not found" {
			log.Warn().Int("Comment ID", id).Msg("Comment not found")
			writeErrorResponse(w, http.StatusNotFound, "Comment not found")
			return
		}
		log.Error().Err(err).Msg("Failed to get comment")
		writeErrorResponse(w, http.StatusInternalServerError, "Failed to get comment")
		return
	}

	// Verify user owns the comment
	if existingComment.UserId != user.ID {
		log.Warn().Int("User ID", user.ID).Int("Comment ID", existingComment.CommentId).Msg("User does not own this comment")
		writeErrorResponse(w, http.StatusForbidden, "You can only update comments you own")
		return
	}

	// Parse the request body
	var req struct {
		Content string `json:"content"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Error().Err(err).Msg("Invalid request body")
		writeErrorResponse(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Validate input
	if req.Content == "" {
		log.Warn().Msg("Missing required field: content")
		writeErrorResponse(w, http.StatusBadRequest, "Content is required")
		return
	}

	// Update comment object with new data
	existingComment.Content = req.Content

	// Call the db to update the comment
	if err := h.db.UpdateComment(existingComment); err != nil {
		log.Error().Err(err).Msg("Failed to update comment")
		writeErrorResponse(w, http.StatusInternalServerError, "Failed to update comment")
		return
	}

	// Success
	log.Info().Int("Comment ID", id).Msg("Successfully updated comment")
	writeJSONResponse(w, http.StatusOK, existingComment)
}

// DELETE /api/comments/{commentId} - Delete a comment
func (h *Handler) DeleteComment(w http.ResponseWriter, r *http.Request) {
	log.Info().Msg("DELETE /api/comments/{commentId} - Deleting comment")

	// Verify user authentification
	username := middleware.GetUsername(r)
	if username == "" {
		log.Warn().Msg("No username in context")
		writeErrorResponse(w, http.StatusUnauthorized, "Unauthorized user")
		return
	}

	// Get user from database
	user, err := h.db.GetUserByUsername(username)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get user info")
		writeErrorResponse(w, http.StatusInternalServerError, "Failed to get user info")
		return
	}

	// Get string commentID from URL
	vars := mux.Vars(r)
	idStr := vars["commentId"]

	// Convert string ID to int
	id, err := strconv.Atoi(idStr)
	if err != nil {
		log.Warn().Str("Comment ID", idStr).Msg("Invalid comment ID format")
		writeErrorResponse(w, http.StatusBadRequest, "Invalid comment ID format")
		return
	}

	// Get existing comment from db
	existingComment, err := h.db.GetCommentById(id)
	if err != nil {
		if err.Error() == "comment not found" {
			log.Warn().Int("Comment ID", id).Msg("Comment not found")
			writeErrorResponse(w, http.StatusNotFound, "Comment not found")
			return
		}
	}

	// Verify comment belongs to user or user deleting is admin
	if existingComment.UserId != user.ID && user.Role != "admin" {
		log.Warn().Int("Comment ID", id).Int("User ID", user.ID).Msg("User does not own this comment")
		writeErrorResponse(w, http.StatusForbidden, "You can only delete your comments")
		return
	}

	// Call db to delete the comment
	if err := h.db.DeleteComment(existingComment.CommentId); err != nil {
		log.Error().Err(err).Msg("Failed to delete comment")
		writeErrorResponse(w, http.StatusInternalServerError, "You can only delete your own comments")
		return
	}

	// Success
	log.Info().Int("Comment ID", id).Msg("Successfully deleted comment")
	writeJSONResponse(w, http.StatusOK, map[string]string{"message": "comment successfully deleted"})
}

// #endregion

// #region Post handlers

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

	log.Info().Int("Post ID", id).Msg("Successfully retrieved post by ID")
	writeJSONResponse(w, http.StatusOK, post)
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

// #endregion

// #region Profile handlers

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

// #endregion

// #region Handler for Users

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

// #endregion

// #region Notifications

// GET /api/notifications - Get all notifications for a user by User ID
func (h *Handler) GetNotificationsByUserId(w http.ResponseWriter, r *http.Request) {
	log.Info().Msg("Getting users notifications")

	// Get username
	username := middleware.GetUsername(r)
	if username == "" {
		log.Warn().Msg("No username in that context")
		writeErrorResponse(w, http.StatusUnauthorized, "Unauthorized user")
		return
	}

	// Get user info from the db
	user, err := h.db.GetUserByUsername(username)
	if err != nil {
		log.Error().Err(err).Msg("failed to get user")
		writeErrorResponse(w, http.StatusInternalServerError, "Failed to get user info")
		return
	}

	notifications, err := h.db.GetNotificationsByUserId(user.ID)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get notifications")
		writeErrorResponse(w, http.StatusInternalServerError, "Failed to get notifications")
		return
	}

	log.Info().Int("Notification Count", len(notifications)).Msg("Successfully retrieved users notifications")
	writeJSONResponse(w, http.StatusOK, notifications)
}

// PUT /api/notifications/{notificationId}/read - Updating notification as read
func (h *Handler) MarkNotificationAsRead(w http.ResponseWriter, r *http.Request) {
	log.Info().Msg("Marking notification as read")

	// Get notif ID from the URL
	vars := mux.Vars(r)
	idStr := vars["notificationId"]

	id, err := strconv.Atoi(idStr)
	if err != nil {
		log.Warn().Str("Notification ID", idStr).Msg("Failed to convert ID into an int")
		writeErrorResponse(w, http.StatusBadRequest, "Invalid Notification ID")
		return
	}

	if err := h.db.MarkNotificationAsRead(id); err != nil {
		log.Error().Err(err).Int("Notification ID", id).Msg("Failed to mark notification as read")
		writeErrorResponse(w, http.StatusInternalServerError, "Failed to update notification")
		return
	}

	log.Info().Int("Notification ID", id).Msg("Successfully marked notification as read")
	writeJSONResponse(w, http.StatusOK, map[string]string{"message": "Successfully marked notification as read"})
}

// Helper function for creating a notification
func (h *Handler) createNotification(comment model.Comment) {
	log.Info().Int("Comment ID", comment.CommentId).Msg("Creating notification for comment")

	// Get the post data
	post, err := h.db.GetPostById(comment.PostId)
	if err != nil {
		log.Error().Err(err).Int("Post ID", comment.PostId).Msg("Failed to get post for notification")
		return
	}

	if post.UserId == comment.UserId {
		return
	}

	notif := model.Notification{
		UserId:    post.UserId,
		PostId:    comment.PostId,
		Message:   fmt.Sprintf("%s commented on your post!", comment.Author),
		CommentId: &comment.CommentId,
	}

	if err := h.db.CreateNotification(&notif); err != nil {
		log.Error().Err(err).Int("Comment ID", comment.CommentId).Msg("Failed to create notification")
	}
}

// #endregion

// #region Post/Comment reaction handlers

// POST /api/posts/{postId}/react - Like or dislike a post
func (h *Handler) ReactToPost(w http.ResponseWriter, r *http.Request) {
	log.Info().Msg("POST /api/posts/{postId}/react - Reacting to a post")

	username := middleware.GetUsername(r)
	if username == "" {
		log.Warn().Msg("No username in the context")
		writeErrorResponse(w, http.StatusUnauthorized, "Not authenticated")
		return
	}

	user, err := h.db.GetUserByUsername(username)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get user")
		writeErrorResponse(w, http.StatusInternalServerError, "Failed to get user")
		return
	}

	vars := mux.Vars(r)
	postId, err := strconv.Atoi(vars["postId"])
	if err != nil {
		log.Warn().Str("Post ID", vars["postId"]).Msg("Invalid post ID format")
		writeErrorResponse(w, http.StatusBadRequest, "Invalid post ID")
		return
	}

	var req struct {
		Reaction string `json:"reaction"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Warn().Err(err).Msg("Invalid request body")
		writeErrorResponse(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.Reaction != "like" && req.Reaction != "dislike" {
		log.Warn().Str("Reaction", req.Reaction).Msg("Invalid reaction value")
		writeErrorResponse(w, http.StatusBadRequest, "Reaction must be 'like' or 'dislike'")
		return
	}

	counts, err := h.db.GetPostReactionCounts(postId, user.ID)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get post reaction counts")
		writeErrorResponse(w, http.StatusInternalServerError, "Failed to get reaction counts")
		return
	}

	reaction := model.Reaction {
		UserId: user.ID,
		TargetId: postId,
		Reaction: req.Reaction,
	}

	// Same reaction clicked twice (toggle off)
	if counts.UserReaction != nil && *counts.UserReaction == req.Reaction {
		if err := h.db.DeletePostReaction(reaction); err != nil {
			log.Error().Err(err).Msg("Failed to delete post reaction")
			writeErrorResponse(w, http.StatusInternalServerError, "Failed to remove reaction")
			return
		}
	} else {
		// Insert or switch reaction
		if err := h.db.UpsertPostReaction(reaction); err != nil {
			log.Error().Err(err).Msg("Failed to upsert post reaction")
			writeErrorResponse(w, http.StatusInternalServerError, "Failed to save reaction")
			return
		}
	}

	// Return updated counts
	updatedCounts, err := h.db.GetPostReactionCounts(postId, user.ID)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get updated post reaction counts")
		writeErrorResponse(w, http.StatusInternalServerError, "Failed to get updated counts")
		return
	}

	log.Info().Int("Post ID", postId).Str("Reaction", req.Reaction).Msg("Successfully reacted to post")
	writeJSONResponse(w, http.StatusOK, updatedCounts)
}

// POST /api/comments/{commentId}/react - Like or dislike a comment
func (h *Handler) ReactToComment(w http.ResponseWriter, r *http.Request) {
	log.Info().Msg("POST /api/comments/{commentId}/react - Reacting to a comment")

	username := middleware.GetUsername(r)
	if username == "" {
		log.Warn().Msg("No username in the context")
		writeErrorResponse(w, http.StatusUnauthorized, "Not authenticated")
		return
	}

	user, err := h.db.GetUserByUsername(username)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get user")
		writeErrorResponse(w, http.StatusInternalServerError, "Failed to get user")
		return
	}

	vars := mux.Vars(r)
	commentId, err := strconv.Atoi(vars["commentId"])
	if err != nil {
		log.Warn().Str("Comment ID", vars["commentId"]).Msg("Invalid comment ID format")
		writeErrorResponse(w, http.StatusBadRequest, "Invalid comment ID")
		return
	}

	var req struct {
		Reaction string `json:"reaction"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Warn().Err(err).Msg("Invalid request body")
		writeErrorResponse(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.Reaction != "like" && req.Reaction != "dislike" {
		log.Warn().Str("Reaction", req.Reaction).Msg("Invalid reaction value")
		writeErrorResponse(w, http.StatusBadRequest, "Reaction must be 'like' or 'dislike'")
		return
	}

	counts, err := h.db.GetCommentReactionCounts(commentId, user.ID)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get comment reaction counts")
		writeErrorResponse(w, http.StatusInternalServerError, "Failed to get reaction counts")
		return
	}

	reaction := model.Reaction {
		UserId: user.ID,
		TargetId: commentId,
		Reaction: req.Reaction,
	}

	// Same reaction clicked twice (toggle off)
	if counts.UserReaction != nil && *counts.UserReaction == req.Reaction {
		if err := h.db.DeleteCommentReaction(reaction); err != nil {
			log.Error().Err(err).Msg("Failed to delete comment reaction")
			writeErrorResponse(w, http.StatusInternalServerError, "Failed to remove reaction")
			return
		}
	} else {
		// Insert or switch reaction
		if err := h.db.UpsertCommentReaction(reaction); err != nil {
			log.Error().Err(err).Msg("Failed to upsert comment reaction")
			writeErrorResponse(w, http.StatusInternalServerError, "Failed to save reaction")
			return
		}
	}

	// Return updated counts
	updatedCounts, err := h.db.GetCommentReactionCounts(commentId, user.ID)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get updated comment reaction counts")
		writeErrorResponse(w, http.StatusInternalServerError, "Failed to get updated counts")
		return
	}

	log.Info().Int("Comment ID", commentId).Str("Reaction", req.Reaction).Msg("Successfully reacted to comment")
	writeJSONResponse(w, http.StatusOK, updatedCounts)
}
