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

	username := middleware.GetUsername(r)
	if username == "" {
		log.Warn().Msg("Failed to get username in the context")
		writeErrorResponse(w, http.StatusUnauthorized, "Unauthorized user")
		return
	}

	user, err := h.db.GetUserByUsername(username)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get user info")
		writeErrorResponse(w, http.StatusInternalServerError, "Failed to get user")
		return
	}

	var response []model.CommentResponse
	for _, c := range comments {
		counts, err := h.db.GetCommentReactionCounts(c.CommentId, user.ID)
		if err != nil {
			log.Error().Err(err).Int("Comment ID", c.CommentId).Msg("Failed to get comment reaction counts")
			writeErrorResponse(w, http.StatusInternalServerError, "Failed to get reaction counts")
			return
		}

		response = append(response, model.CommentResponse{Comment: c, Reactions: counts})
	}

	log.Info().Int("count", len(comments)).Msg("Successfully retrieved comments on post")
	writeJSONResponse(w, http.StatusOK, response)

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
