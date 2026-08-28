-- Preserve sub-second ordering and add the ID tie-breaker used by opaque
-- pagination cursors. Run this once on databases created before this change.
ALTER TABLE posts
  MODIFY created_at DATETIME(6) NOT NULL,
  DROP INDEX idx_posts_author_created,
  ADD KEY idx_posts_author_created (author_id, created_at, id),
  DROP INDEX idx_posts_created_at,
  ADD KEY idx_posts_created_at (created_at, id);

ALTER TABLE comments
  MODIFY created_at DATETIME(6) NOT NULL,
  DROP INDEX idx_comments_post_created,
  ADD KEY idx_comments_post_created (post_id, created_at, id),
  DROP INDEX idx_comments_user_created,
  ADD KEY idx_comments_user_created (user_id, created_at, id);

ALTER TABLE notifications
  MODIFY created_at DATETIME(6) NOT NULL,
  DROP INDEX idx_notifications_user_read_created,
  ADD KEY idx_notifications_user_read_created (user_id, is_read, created_at, id),
  DROP INDEX idx_notifications_user_created,
  ADD KEY idx_notifications_user_created (user_id, created_at, id);
