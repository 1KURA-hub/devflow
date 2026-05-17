package model

import "time"

type ProcessedEvent struct {
	EventID   string    `gorm:"primaryKey;type:varchar(64)" json:"event_id"`
	EventType string    `gorm:"type:varchar(64);not null" json:"event_type"`
	CreatedAt time.Time `json:"created_at"`
}

func (ProcessedEvent) TableName() string {
	return "processed_events"
}
