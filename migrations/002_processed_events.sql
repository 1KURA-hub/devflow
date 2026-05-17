CREATE TABLE IF NOT EXISTS processed_events (
  event_id VARCHAR(64) PRIMARY KEY,
  event_type VARCHAR(64) NOT NULL,
  created_at DATETIME NOT NULL
);
