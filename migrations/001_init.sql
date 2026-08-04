CREATE TABLE IF NOT EXISTS users (
  id BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
  username VARCHAR(64) NOT NULL,
  password_hash VARCHAR(255) NOT NULL,
  nickname VARCHAR(64) NOT NULL,
  bio VARCHAR(255) NOT NULL DEFAULT '',
  avatar_url VARCHAR(512) NOT NULL DEFAULT '',
  status TINYINT NOT NULL DEFAULT 1,
  created_at DATETIME(6) NOT NULL,
  updated_at DATETIME(6) NOT NULL,
  deleted_at DATETIME NULL,
  UNIQUE KEY uk_users_username (username),
  KEY idx_users_deleted_at (deleted_at)
);

CREATE TABLE IF NOT EXISTS posts (
  id BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
  author_id BIGINT UNSIGNED NOT NULL,
  title VARCHAR(120) NOT NULL,
  content TEXT NOT NULL,
  tags VARCHAR(255) NOT NULL DEFAULT '',
  like_count BIGINT NOT NULL DEFAULT 0,
  comment_count BIGINT NOT NULL DEFAULT 0,
  favorite_count BIGINT NOT NULL DEFAULT 0,
  status TINYINT NOT NULL DEFAULT 1,
  created_at DATETIME(6) NOT NULL,
  updated_at DATETIME(6) NOT NULL,
  deleted_at DATETIME NULL,
  KEY idx_posts_author_created (author_id, created_at, id),
  KEY idx_posts_created_at (created_at, id),
  KEY idx_posts_deleted_at (deleted_at)
);

CREATE TABLE IF NOT EXISTS follows (
  id BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
  follower_id BIGINT UNSIGNED NOT NULL,
  followee_id BIGINT UNSIGNED NOT NULL,
  created_at DATETIME(6) NOT NULL,
  UNIQUE KEY uk_follows_pair (follower_id, followee_id),
  KEY idx_follows_follower (follower_id),
  KEY idx_follows_followee (followee_id)
);

CREATE TABLE IF NOT EXISTS likes (
  id BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
  user_id BIGINT UNSIGNED NOT NULL,
  post_id BIGINT UNSIGNED NOT NULL,
  created_at DATETIME(6) NOT NULL,
  UNIQUE KEY uk_likes_user_post (user_id, post_id),
  KEY idx_likes_user (user_id),
  KEY idx_likes_post (post_id)
);

CREATE TABLE IF NOT EXISTS favorites (
  id BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
  user_id BIGINT UNSIGNED NOT NULL,
  post_id BIGINT UNSIGNED NOT NULL,
  created_at DATETIME(6) NOT NULL,
  UNIQUE KEY uk_favorites_user_post (user_id, post_id),
  KEY idx_favorites_user (user_id),
  KEY idx_favorites_post (post_id)
);

CREATE TABLE IF NOT EXISTS comments (
  id BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
  post_id BIGINT UNSIGNED NOT NULL,
  user_id BIGINT UNSIGNED NOT NULL,
  content VARCHAR(1000) NOT NULL,
  status TINYINT NOT NULL DEFAULT 1,
  created_at DATETIME(6) NOT NULL,
  updated_at DATETIME(6) NOT NULL,
  deleted_at DATETIME NULL,
  KEY idx_comments_post_created (post_id, created_at, id),
  KEY idx_comments_user_created (user_id, created_at, id),
  KEY idx_comments_deleted_at (deleted_at)
);

CREATE TABLE IF NOT EXISTS notifications (
  id BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
  user_id BIGINT UNSIGNED NOT NULL,
  actor_id BIGINT UNSIGNED NOT NULL,
  type VARCHAR(32) NOT NULL,
  post_id BIGINT UNSIGNED NULL,
  comment_id BIGINT UNSIGNED NULL,
  content VARCHAR(255) NOT NULL,
  is_read BOOLEAN NOT NULL DEFAULT FALSE,
  created_at DATETIME(6) NOT NULL,
  KEY idx_notifications_user_read_created (user_id, is_read, created_at, id),
  KEY idx_notifications_user_created (user_id, created_at, id),
  KEY idx_notifications_post (post_id),
  KEY idx_notifications_comment (comment_id)
);
