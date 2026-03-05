package model

import "time"

type Comment struct {
	CommentId  int       `json:"comment_id" db:"comment_id"`
	UserId     int       `json:"user_id" db:"user_id"`
	PostId     int       `json:"post_id" db:"post_id"`
	Content    string    `json:"content" db:"content"`
	Author     string    `json:"author" db:"author"`
	DatePosted time.Time `json:"date_posted" db:"date_posted"`
}

type Post struct {
	PostId     int       `json:"post_id" db:"post_id"`
	UserId     int       `json:"user_id" db:"user_id"`
	Title      string    `json:"title" db:"title"`
	Content    string    `json:"content" db:"content"`
	Author     string    `json:"author" db:"author"`
	DatePosted time.Time `json:"date_posted" db:"date_posted"`
}

type Profile struct {
	UserId         int       `json:"user_id" db:"user_id"`
	FirstName      string    `json:"first_name" db:"first_name"`
	LastName       string    `json:"last_name" db:"last_name"`
	Email          string    `json:"email" db:"email"`
	GithubLink     string    `json:"github_link" db:"github_link"`
	City           string    `json:"city" db:"city"`
	State          string    `json:"state" db:"state"`
	DateRegistered time.Time `json:"date_registered" db:"date_registered"`
}

type User struct {
	ID             int    `json:"user_id" db:"user_id"`
	Username       string `json:"username" db:"username"`
	HashedPassword string `json:"-" db:"hashed_password"`
	Role           string `json:"role" db:"role"`
	FirstName      string `json:"first_name" db:"first_name"`
	LastName       string `json:"last_name" db:"last_name"`
}

type Notification struct {
	ID        int       `json:"notification_id" db:"notification_id"`
	UserId    int       `json:"user_id" db:"user_id"`
	PostId    int       `json:"post_id" db:"post_id"`
	Message   string    `json:"message" db:"message"`
	IsRead    bool      `json:"is_read" db:"is_read"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	CommentId *int      `json:"comment_id" db:"comment_id"`
}

type Reaction struct {
	UserId   int    `json:"user_id" db:"user_id"`
	TargetId int    `json:"target_id"`
	Reaction string `json:"reaction" db:"reaction"`
}

type ReactionCounts struct {
	Likes        int     `json:"likes" db:"likes"`
	Dislikes     int     `json:"dislikes" db:"dislikes"`
	UserReaction *string `json:"user_reaction" db:"user_reaction"`
}

type PostResponse struct {
	Post
	Reactions ReactionCounts `json:"reactions"`
}

type CommentResponse struct {
	Comment
	Reactions ReactionCounts `json:"reactions"`
}
