package handler

import (
	"byte-board/internal/middleware"
	"byte-board/internal/model"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/rs/zerolog/log"
)

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

// Helper function for creating a notification
func (h *Handler) createReactionNotification(comment *model.Comment, post *model.Post, reactor, reaction string) {
	log.Info().Msg("Creating notification")

	var ownerId, postId int
	var commentId *int
	var target string

	if comment == nil {
		// Post reaction
		ownerId = post.UserId
		postId = post.PostId
		target = "post"
	} else {
		// Comment reaction
		ownerId = comment.UserId
		postId = comment.PostId
		commentId = &comment.CommentId
		target = "comment"
	}

	notif := model.Notification{
		UserId:    ownerId,
		PostId:    postId,
		CommentId: commentId,
		Message:   fmt.Sprintf("%s %sd your %s", reactor, reaction, target),
	}

	if err := h.db.CreateNotification(&notif); err != nil {
		log.Error().Err(err).Msg("Failed to create reaction notification")
	}
}
