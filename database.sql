-- ----------------------------------------------------------------------
-- Target DBMS:           PostgreSQL
-- Project name:          ByteBoard
-- ----------------------------------------------------------------------
--
-- Note: Run this script while connected to the byteboard_db database
-- Command: psql -U postgres -d byteboard_db -f database.sql
-- Or: cat database.sql | docker exec -i byte-db psql -U postgres -d byteboard_db
-- ----------------------------------------------------------------------

-- Drop tables if they exist
DROP TABLE IF EXISTS comment_reactions CASCADE;

DROP TABLE IF EXISTS post_reactions CASCADE;

DROP TABLE IF EXISTS notifications CASCADE;

DROP TABLE IF EXISTS comments CASCADE;

DROP TABLE IF EXISTS posts CASCADE;

DROP TABLE IF EXISTS profiles CASCADE;

DROP TABLE IF EXISTS users CASCADE;

-- ----------------------------------------------------------------------
-- Tables
-- ----------------------------------------------------------------------

-- Creating tables
CREATE TABLE users (
    user_id SERIAL PRIMARY KEY,
    username VARCHAR(50) NOT NULL UNIQUE,
    hashed_password VARCHAR(255) NOT NULL,
    role VARCHAR(50) NOT NULL,
    first_name VARCHAR(50), -- ADD THIS
    last_name VARCHAR(50) -- ADD THIS
);

CREATE TABLE profiles (
    user_id INTEGER PRIMARY KEY,
    first_name VARCHAR(50),
    last_name VARCHAR(50),
    email VARCHAR(200),
    github_link VARCHAR(75),
    city VARCHAR(50),
    state VARCHAR(50),
    date_registered DATE,
    FOREIGN KEY (user_id) REFERENCES users (user_id) ON DELETE CASCADE
);

CREATE TABLE posts (
    post_id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL,
    title VARCHAR(255) NOT NULL,
    content TEXT NOT NULL,
    author VARCHAR(50) NOT NULL,
    date_posted TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users (user_id) ON DELETE CASCADE
);

CREATE TABLE comments (
    comment_id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL,
    post_id INTEGER NOT NULL,
    content TEXT NOT NULL,
    author VARCHAR(50) NOT NULL,
    date_posted TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users (user_id) ON DELETE CASCADE,
    FOREIGN KEY (post_id) REFERENCES posts (post_id) ON DELETE CASCADE
);

CREATE TABLE notifications (
    notification_id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL,
    post_id INTEGER NOT NULL,
    message VARCHAR(255) NOT NULL,
    is_read BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    comment_id INTEGER,
    FOREIGN KEY (user_id) REFERENCES users (user_id) ON DELETE CASCADE,
    FOREIGN KEY (post_id) REFERENCES posts (post_id) ON DELETE CASCADE,
    FOREIGN KEY (comment_id) REFERENCES comments (comment_id) ON DELETE CASCADE
);

CREATE TABLE post_reactions (
    reaction_id SERIAL PRIMARY KEY,
    user_id     INTEGER NOT NULL REFERENCES users(user_id) ON DELETE CASCADE,
    post_id     INTEGER NOT NULL REFERENCES posts(post_id) ON DELETE CASCADE,
    reaction    VARCHAR(10) NOT NULL CHECK (reaction IN ('like', 'dislike')),
    CONSTRAINT unique_post_reaction UNIQUE (user_id, post_id)
);

CREATE TABLE comment_reactions (
    reaction_id SERIAL PRIMARY KEY,
    user_id     INTEGER NOT NULL REFERENCES users(user_id) ON DELETE CASCADE,
    comment_id  INTEGER NOT NULL REFERENCES comments(comment_id) ON DELETE CASCADE,
    reaction    VARCHAR(10) NOT NULL CHECK (reaction IN ('like', 'dislike')),
    CONSTRAINT unique_comment_reaction UNIQUE (user_id, comment_id)
);

-- Create indexes for better query performance
CREATE INDEX idx_posts_user_id ON posts (user_id);

CREATE INDEX idx_posts_date_posted ON posts (date_posted);

CREATE INDEX idx_comments_post_id ON comments (post_id);

CREATE INDEX idx_comments_user_id ON comments (user_id);

CREATE INDEX idx_notifications_user_id ON notifications (user_id);

CREATE INDEX idx_notifications_is_read ON notifications (user_id, is_read);

CREATE INDEX idx_post_reactions_post_id ON post_reactions (post_id);

CREATE INDEX idx_comment_reactions_comment_id ON comment_reactions (comment_id);