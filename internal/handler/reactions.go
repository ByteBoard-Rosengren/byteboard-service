package handler

import (
	"byte-board/internal/middleware"
	"byte-board/internal/model"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/rs/zerolog/log"
)

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

	post, err := h.db.GetPostById(postId)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get post by ID")
		writeErrorResponse(w, http.StatusInternalServerError, "Failed to get post")
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

	reaction := model.Reaction{
		UserId:   user.ID,
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
		if post.UserId != user.ID {
			go h.db.DeleteReactionNotification(post.UserId, postId, nil)
		}
	} else {
		// Insert or switch reaction
		if err := h.db.UpsertPostReaction(reaction); err != nil {
			log.Error().Err(err).Msg("Failed to upsert post reaction")
			writeErrorResponse(w, http.StatusInternalServerError, "Failed to save reaction")
			return
		}
		if post.UserId != user.ID {
			prevReaction := counts.UserReaction
			go func() {
				if prevReaction != nil {
					// Switch reaction - delete old notification before making new one
					h.db.DeleteReactionNotification(post.UserId, postId, nil)
				}
				h.createReactionNotification(nil, post, user.Username, req.Reaction)
			}()
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

	comment, err := h.db.GetCommentById(commentId)
	if err != nil {
		log.Error().Err(err).Msg("Could not get comment by ID")
		writeErrorResponse(w, http.StatusInternalServerError, "failed to get comment")
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

	reaction := model.Reaction{
		UserId:   user.ID,
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
		if comment.UserId != user.ID {
			go h.db.DeleteReactionNotification(comment.UserId, comment.PostId, &comment.CommentId)
		}
	} else {
		// Insert or switch reaction
		if err := h.db.UpsertCommentReaction(reaction); err != nil {
			log.Error().Err(err).Msg("Failed to upsert comment reaction")
			writeErrorResponse(w, http.StatusInternalServerError, "Failed to save reaction")
			return
		}
		if comment.UserId != user.ID {
			prevReaction := counts.UserReaction
			go func() {
				if prevReaction != nil {
					// Switch reaction - delete old notif before making new one
					h.db.DeleteReactionNotification(comment.UserId, comment.PostId, &commentId)
				}
				h.createReactionNotification(comment, nil, user.Username, req.Reaction)
			}()
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
