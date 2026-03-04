package repository

import (
	"byte-board/internal/appconfig"
	"byte-board/internal/model"
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"
	"github.com/rs/zerolog/log"
)

type DB struct {
	*sql.DB
}

// Create new database connection
func New(cfg *appconfig.Config) (*DB, error) {
	// Get the database URL
	databaseURL, err := cfg.GetDatabaseURL()
	if err != nil {
		return nil, fmt.Errorf("could not get the database url: %w", err)
	}

	// Open connection to database
	db, err := sql.Open("postgres", databaseURL)
	if err != nil {
		return nil, fmt.Errorf("could not establish connection with database: %w", err)
	}

	// Ping database (verify conn to db is still alive)
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	log.Info().Msg("Database successfully connected!")
	return &DB{DB: db}, nil
}

// #region Comments

// Get all comments in the db
func (db *DB) GetAllComments() ([]model.Comment, error) {
	query := "SELECT * FROM comments"

	rows, err := db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query comments: %w", err)
	}
	defer rows.Close()

	var commentsList []model.Comment
	for rows.Next() {
		var comment model.Comment
		err := rows.Scan(&comment.CommentId, &comment.UserId, &comment.PostId, &comment.Content, &comment.Author, &comment.DatePosted)
		if err != nil {
			return nil, fmt.Errorf("failed to scan comments: %w", err)
		}

		commentsList = append(commentsList, comment)
	}

	return commentsList, nil
}

// Get comment by ID
func (db *DB) GetCommentById(commentId int) (*model.Comment, error) {
	query := "SELECT * FROM comments WHERE comment_id = $1"

	var comment model.Comment
	err := db.QueryRow(query, commentId).Scan(&comment.CommentId, &comment.UserId, &comment.PostId, &comment.Content, &comment.Author, &comment.DatePosted)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("comment not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to query comments: %w", err)
	}

	return &comment, nil
}

// Get all comments on a post
func (db *DB) GetCommentsByPost(postId int) ([]model.Comment, error) {
	query := "SELECT * FROM comments WHERE post_id = $1"

	rows, err := db.Query(query, postId)
	if err != nil {
		return nil, fmt.Errorf("failed to query comments on post: %w", err)
	}
	defer rows.Close()

	var commentList []model.Comment
	for rows.Next() {
		var comment model.Comment
		err := rows.Scan(&comment.CommentId, &comment.UserId, &comment.PostId, &comment.Content, &comment.Author, &comment.DatePosted)
		if err != nil {
			return nil, fmt.Errorf("failed to scan comments on post")
		}

		commentList = append(commentList, comment)
	}

	return commentList, nil
}

// Create comment on a post
func (db *DB) CreateComment(comment *model.Comment, postId int) error {
	log.Info().Int("PostID", postId).Msg("Creating comment on post")

	query := `
		INSERT INTO comments (user_id, post_id, content, author, date_posted)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING comment_id
			`

	err := db.QueryRow(query, comment.UserId, comment.PostId, comment.Content, comment.Author, comment.DatePosted).
		Scan(&comment.CommentId)
	if err != nil {
		return fmt.Errorf("failed to create comment: %w", err)
	}

	return nil
}

// Update a comment
func (db *DB) UpdateComment(comment *model.Comment) error {
	log.Info().Int("ID", comment.CommentId).Msg("Updating comment in the database")

	query := `
		UPDATE comments 
		SET content = $2, 
		author = $3 
		WHERE comment_id = $1
	`

	result, err := db.Exec(query, comment.CommentId, comment.Content, comment.Author)
	if err != nil {
		return fmt.Errorf("failed to update comment: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("comment not found")
	}

	return nil
}

// Delete a comment
func (db *DB) DeleteComment(id int) error {
	log.Info().Int("ID", id).Msg("Deleting comment from the database")

	query := "DELETE FROM comments WHERE comment_id = $1"

	result, err := db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to delete comment: %w", err)
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("comment not found")
	}

	return nil
}

// #endregion

// #region Posts

// Get all posts in the DB
func (db *DB) GetAllPosts() ([]model.Post, error) {
	query := "SELECT * FROM posts ORDER BY date_posted DESC"

	rows, err := db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query rows: %w", err)
	}
	defer rows.Close()

	var postList []model.Post
	for rows.Next() {
		var post model.Post
		err := rows.Scan(&post.PostId, &post.UserId, &post.Title, &post.Content, &post.Author, &post.DatePosted)
		if err != nil {
			return nil, fmt.Errorf("failed to scan rows: %w", err)
		}

		postList = append(postList, post)
	}

	return postList, nil
}

// Get post by post ID
func (db *DB) GetPostById(postId int) (*model.Post, error) {
	query := "SELECT * FROM posts WHERE post_id = $1"

	var post model.Post
	err := db.QueryRow(query, postId).Scan(&post.PostId, &post.UserId, &post.Title, &post.Content, &post.Author, &post.DatePosted)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("post not found")
	}

	return &post, nil
}

// Get all posts made by a user
func (db *DB) GetPostsByUserId(userId int) ([]model.Post, error) {
	query := "SELECT * FROM posts WHERE user_id = $1"

	rows, err := db.Query(query, userId)
	if err != nil {
		return nil, fmt.Errorf("failed to query rows: %w", err)
	}

	var postList []model.Post
	for rows.Next() {
		var post model.Post
		err := rows.Scan(&post.PostId, &post.UserId, &post.Title, &post.Content, &post.Author, &post.DatePosted)
		if err != nil {
			return nil, fmt.Errorf("failed to scan rows: %w", err)
		}

		postList = append(postList, post)
	}

	if len(postList) == 0 {
		return nil, fmt.Errorf("users posts not found")
	}
	return postList, nil
}

// POST api/posts - Create a post
func (db *DB) CreatePost(post *model.Post) error {
	query := `
		INSERT INTO posts (user_id, title, content, author, date_posted) 
		VALUES ($1, $2, $3, $4, $5) 
		RETURNING post_id
	`

	err := db.QueryRow(query, post.UserId, post.Title, post.Content, post.Author, post.DatePosted).
		Scan(&post.PostId)
	if err != nil {
		return fmt.Errorf("failed to create post: %w", err)
	}

	return nil
}

// PUT api/posts/{postId} - Update a post
func (db *DB) UpdatePost(post *model.Post) error {
	query := `
		UPDATE posts
		SET user_id = $2, title = $3, content = $4, author = $5, date_posted = $6
		WHERE post_id = $1
	`

	result, err := db.Exec(query, post.PostId, post.UserId, post.Title, post.Content, post.Author, post.DatePosted)
	if err != nil {
		return fmt.Errorf("failed to update post: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	log.Info().Int("post_id", post.PostId).Int64("rows affected", rowsAffected).Msg("Post update query executed")

	if rowsAffected == 0 {
		log.Warn().Int("post_id", post.PostId).Msg("No rows affected - post not found")
	}

	log.Info().Int("post_id", post.PostId).Msg("Successfully updated post in database")
	return nil
}

// DELETE api/posts/{postId} - Delete a post
func (db *DB) DeletePost(postId int) error {
	log.Info().Int("ID", postId).Msg("Deleting post from the database")

	query := "DELETE FROM posts WHERE post_id = $1"
	result, err := db.Exec(query, postId)
	if err != nil {
		log.Error().Err(err).Int("PostID", postId).Msg("Failed to execute post deletion query")
		return fmt.Errorf("failed to delete post: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	log.Info().Int("PostID", postId).Int64("rows affected", rowsAffected).Msg("Post deletion query executed")

	if rowsAffected == 0 {
		log.Warn().Int("PostID", postId).Msg("No rows affected - post not found")
		return fmt.Errorf("post not found")
	}

	log.Info().Int("PostID", postId).Msg("Successfully deleted post from the database")
	return nil
}

// #endregion

// #region Profiles

// Get all profiles
func (db *DB) GetAllProfiles() ([]model.Profile, error) {
	query := "SELECT * FROM profiles"

	rows, err := db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query profiles: %w", err)
	}

	var profileList []model.Profile
	for rows.Next() {
		var profile model.Profile
		err := rows.Scan(&profile.UserId, &profile.FirstName, &profile.LastName, &profile.Email, &profile.GithubLink, &profile.City, &profile.State, &profile.DateRegistered)
		if err != nil {
			return nil, fmt.Errorf("failed to scan profiles: %w", err)
		}

		profileList = append(profileList, profile)
	}

	return profileList, nil
}

// Get profile by User ID
func (db *DB) GetProfileByUserId(userId int) (*model.Profile, error) {
	query := "SELECT * FROM profiles WHERE user_id = $1"

	var profile model.Profile
	err := db.QueryRow(query, userId).Scan(&profile.UserId, &profile.FirstName, &profile.LastName, &profile.Email, &profile.GithubLink, &profile.City, &profile.State, &profile.DateRegistered)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("profile not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to query profiles: %w", err)
	}

	return &profile, err
}

// Create a profile
func (db *DB) CreateProfile(profile *model.Profile) (*model.Profile, error) {
	query := `
		INSERT INTO profiles (user_id, first_name, last_name, email, github_link, city, state, date_registered)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`

	_, err := db.Exec(query,
		profile.UserId,
		profile.FirstName,
		profile.LastName,
		profile.Email,
		profile.GithubLink,
		profile.City,
		profile.State,
		profile.DateRegistered)
	if err != nil {
		return nil, fmt.Errorf("failed to create profile: %w", err)
	}

	return profile, nil
}

// Update a profile
func (db *DB) UpdateProfile(profile *model.Profile) error {
	log.Info().Int("User ID:", profile.UserId).Msg("Updating user profile in the db")

	query := `
		UPDATE profiles 
		SET first_name = $2,
		last_name = $3,
		email = $4,
		github_link = $5,
		city = $6,
		state = $7
		WHERE user_id = $1
	`

	// Execute query
	result, err := db.Exec(query, profile.UserId, profile.FirstName, profile.LastName, profile.Email, profile.GithubLink, profile.City, profile.State)
	if err != nil {
		return fmt.Errorf("failed to update users profile: %w", err)
	}

	// Get rows affected
	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	log.Info().Int("User ID", profile.UserId).Int64("Rows affected", rows).Msg("Profile update query was executed")

	// Verify profile exists
	if rows == 0 {
		return fmt.Errorf("profile not found")
	}

	return nil
}

// Delete a profile
func (db *DB) DeleteProfile(userId int) error {
	log.Info().Int("User ID", userId).Msg("Deleting user's profile")

	query := "DELETE FROM profiles WHERE user_id = $1"
	result, err := db.Exec(query, userId)
	if err != nil {
		return fmt.Errorf("Failed to delete profile: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("profile not found")
	}

	return nil
}

// #endregion

// #region Users

// Get all users
func (db *DB) GetAllUsers() ([]model.User, error) {
	query := "SELECT * FROM users"

	rows, err := db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query users")
	}

	var userList []model.User
	for rows.Next() {
		var user model.User
		err := rows.Scan(&user.ID, &user.Username, &user.HashedPassword, &user.Role, &user.FirstName, &user.LastName)
		if err != nil {
			return nil, fmt.Errorf("failed to scan users")
		}

		userList = append(userList, user)
	}

	return userList, nil
}

// Get user by user ID
func (db *DB) GetUserByID(userId int) (*model.User, error) {
	query := "SELECT * FROM users WHERE user_id = $1"

	var user model.User
	err := db.QueryRow(query, userId).Scan(&user.ID, &user.Username, &user.HashedPassword, &user.Role, &user.FirstName, &user.LastName)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("user not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to query or scan rows: %w", err)
	}

	return &user, nil
}

// GET api/users/username/{username} - Get user by username
func (db *DB) GetUserByUsername(username string) (*model.User, error) {
	query := "SELECT * FROM users WHERE username = $1"

	var user model.User
	err := db.QueryRow(query, username).Scan(&user.ID, &user.Username, &user.HashedPassword, &user.Role, &user.FirstName, &user.LastName)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("username not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to query or scan rows: %w", err)
	}

	return &user, nil
}

// Create new user
func (db *DB) CreateUser(user *model.User) error {
	query := `
		INSERT INTO users (username, hashed_password, role, first_name, last_name)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING user_id
	`

	err := db.QueryRow(query, user.Username, user.HashedPassword, user.Role, user.FirstName, user.LastName).Scan(&user.ID)
	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}

	return nil
}

// Update user
func (db *DB) UpdateUser(user *model.User) error {
	query := `
		UPDATE users
		SET username = $1,
		hashed_password = $2,
		role = $3,
		first_name = $4,
		last_name = $5
		WHERE user_id = $6
	`

	result, err := db.Exec(query, user.Username, user.HashedPassword, user.Role, user.FirstName, user.LastName, user.ID)
	if err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("user not found")
	}

	return nil
}

// Delete user
func (db *DB) DeleteUser(userId int) error {
	query := "DELETE FROM users WHERE user_id = $1"

	result, err := db.Exec(query, userId)
	if err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("user not found")
	}

	return nil
}

// Check if username already exists
func (db *DB) UserExists(username string) (bool, error) {
	query := "SELECT EXISTS(SELECT 1 FROM users WHERE username = $1)"

	var exists bool
	err := db.QueryRow(query, username).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check if user exists: %w", err)
	}

	return exists, nil
}

// #endregion

// #region Notifications

// GET /api/notifications - Get notification by user ID
func (db *DB) GetNotificationsByUserId(userId int) ([]model.Notification, error) {
	query := "SELECT * FROM notifications WHERE user_id = $1 ORDER BY created_at DESC;"

	rows, err := db.Query(query, userId)
	if err != nil {
		return nil, fmt.Errorf("failed to query notifications")
	}
	defer rows.Close()

	var notifications []model.Notification
	for rows.Next() {
		var notif model.Notification
		err := rows.Scan(&notif.ID, &notif.UserId, &notif.PostId, &notif.Message, &notif.IsRead, &notif.CreatedAt, &notif.CommentId)

		if err != nil {
			return nil, fmt.Errorf("failed to scan for notifications: %w", err)
		}

		notifications = append(notifications, notif)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate notifications: %w", err)
	}

	return notifications, nil
}

// PUT /api/notification/{notificationId}/read - Set a notification as read
func (db *DB) MarkNotificationAsRead(notificationId int) error {
	query := `
		UPDATE notifications 
		SET is_read = TRUE 
		WHERE notification_id = $1;
	`

	result, err := db.Exec(query, notificationId)
	if err != nil {
		return fmt.Errorf("failed to update users profile: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("notification not found")
	}

	return nil
}

// Create a new notification
func (db *DB) CreateNotification(notif *model.Notification) error {
	query := `
		INSERT INTO notifications (user_id, post_id, message, comment_id) 
		VALUES ($1, $2, $3, $4)
	`

	_, err := db.Exec(query,
		notif.UserId,
		notif.PostId,
		notif.Message,
		notif.CommentId,
	)
	if err != nil {
		return fmt.Errorf("failed to create notification: %w", err)
	}

	return nil
}

// #endregion

// # region User Reactions

// Post Reactions
func (db *DB) UpsertPostReaction(reaction model.Reaction) error {
	query := `
		INSERT INTO post_reactions (user_id, post_id, reaction) 
		VALUES ($1, $2, $3) 
		ON CONFLICT (user_id, post_id) 
		DO UPDATE SET reaction = EXCLUDED.reaction;
	`

	_, err := db.Exec(query, reaction.UserId, reaction.TargetId, reaction.Reaction)
	if err != nil {
		return fmt.Errorf("failed to upsert post reaction: %w", err)
	}

	return nil
}

// Remove post reaction
func (db *DB) DeletePostReaction(reaction model.Reaction) error {
	query := `
		DELETE FROM post_reactions 
		WHERE user_id = $1 AND post_id = $2;
	`

	result, err := db.Exec(query, reaction.UserId, reaction.TargetId)
	if err != nil {
		return fmt.Errorf("failed to delete post reaction: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("reaction not found")
	}

	return nil
}

// Get the counts of the post likes and dislikes
func (db *DB) GetPostReactionCounts(postId, viewerId int) (model.ReactionCounts, error) {
	query := `
	SELECT
		COUNT(*) FILTER (WHERE reaction = 'like') AS likes,
		COUNT(*) FILTER (WHERE reaction = 'dislike') AS dislikes,
		MAX(reaction) FILTER (WHERE user_id = $2) AS user_reaction
	FROM post_reactions 
	WHERE post_id = $1;
	`

	var count model.ReactionCounts
	err := db.QueryRow(query, postId, viewerId).Scan(
		&count.Likes,
		&count.Dislikes,
		&count.UserReaction,
	)
	if err != nil {
		return model.ReactionCounts{}, fmt.Errorf("failed to get post reaction counts: %w", err)
	}

	return count, nil
}

// Comment Reactions
func (db *DB) UpsertCommentReaction(reaction model.Reaction) error {
	query := `
		INSERT INTO comment_reactions (user_id, comment_id, reaction) 
		VALUES ($1, $2, $3) 
		ON CONFLICT (user_id, comment_id) 
		DO UPDATE SET reaction = EXCLUDED.reaction;
	`

	_, err := db.Exec(query, reaction.UserId, reaction.TargetId, reaction.Reaction)
	if err != nil {
		return fmt.Errorf("failed to upsert comment reaction: %w", err)
	}

	return nil
}

// Remove comment reaction
func (db *DB) DeleteCommentReaction(reaction model.Reaction) error {
	query := `
		DELETE FROM comment_reactions 
		WHERE user_id = $1 AND comment_id = $2;
	`

	result, err := db.Exec(query, reaction.UserId, reaction.TargetId)
	if err != nil {
		return fmt.Errorf("failed to delete comment reaction: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get the rows affected: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("reaction not found")
	}

	return nil
}

// Get the counts of the comment likes and dislikes
func (db *DB) GetCommentReactionCounts(commentId, viewerId int) (model.ReactionCounts, error) {
	query := `
	SELECT
		COUNT(*) FILTER (WHERE reaction = 'like') AS likes,
		COUNT(*) FILTER (WHERE reaction = 'dislike') AS dislikes,
		MAX(reaction) FILTER (WHERE user_id = $2) AS user_reaction
	FROM comment_reactions 
	WHERE comment_id = $1;
	`

	var count model.ReactionCounts
	err := db.QueryRow(query, commentId, viewerId).Scan(
		&count.Likes,
		&count.Dislikes,
		&count.UserReaction,
	)
	if err != nil {
		return model.ReactionCounts{}, fmt.Errorf("failed to get comment reaction counts: %w", err)
	}

	return count, nil
}
// #endregion
