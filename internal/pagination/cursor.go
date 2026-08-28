package pagination

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"time"
)

type Kind string

const (
	KindChronological Kind = "time"
	KindHot           Kind = "hot"
)

var ErrInvalidCursor = errors.New("invalid cursor")

// Cursor is an opaque, versioned API cursor. Chronological lists use
// CreatedAtMicros + ID as a stable tie-breaker; hot lists use Score + ID.
type Cursor struct {
	Version         int    `json:"v"`
	Kind            Kind   `json:"k"`
	CreatedAtMicros int64  `json:"t,omitempty"`
	ID              uint64 `json:"id"`
	Score           int64  `json:"s,omitempty"`
}

func Chronological(createdAt time.Time, id uint64) *Cursor {
	return &Cursor{
		Version:         1,
		Kind:            KindChronological,
		CreatedAtMicros: createdAt.UnixMicro(),
		ID:              id,
	}
}

func Hot(score int64, id uint64) *Cursor {
	return &Cursor{
		Version: 1,
		Kind:    KindHot,
		Score:   score,
		ID:      id,
	}
}

func (c *Cursor) CreatedAt() time.Time {
	if c == nil || c.CreatedAtMicros == 0 {
		return time.Time{}
	}
	return time.UnixMicro(c.CreatedAtMicros)
}

func Encode(cursor *Cursor) (string, error) {
	if cursor == nil {
		return "", ErrInvalidCursor
	}
	if err := validate(cursor, cursor.Kind); err != nil {
		return "", err
	}
	payload, err := json.Marshal(cursor)
	if err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(payload), nil
}

// Decode accepts the current opaque format. For chronological endpoints it
// also accepts the legacy RFC3339 timestamp cursor so in-flight clients can
// finish pagination during a rolling deployment.
func Decode(raw string, expected Kind) (*Cursor, error) {
	if raw == "" {
		return nil, nil
	}

	payload, err := base64.RawURLEncoding.DecodeString(raw)
	if err == nil {
		var cursor Cursor
		if json.Unmarshal(payload, &cursor) == nil && validate(&cursor, expected) == nil {
			return &cursor, nil
		}
	}

	if expected == KindChronological {
		createdAt, legacyErr := time.Parse(time.RFC3339Nano, raw)
		if legacyErr == nil {
			// A legacy cursor has no ID tie-breaker. ID=0 preserves its previous
			// "strictly older timestamp" behavior without duplicating rows.
			return &Cursor{
				Version:         1,
				Kind:            KindChronological,
				CreatedAtMicros: createdAt.UnixMicro(),
			}, nil
		}
	}

	return nil, ErrInvalidCursor
}

func validate(cursor *Cursor, expected Kind) error {
	if cursor == nil || cursor.Version != 1 || cursor.Kind != expected {
		return ErrInvalidCursor
	}
	switch cursor.Kind {
	case KindChronological:
		if cursor.CreatedAtMicros == 0 {
			return ErrInvalidCursor
		}
	case KindHot:
		if cursor.Score <= 0 || cursor.ID == 0 {
			return ErrInvalidCursor
		}
	default:
		return ErrInvalidCursor
	}
	return nil
}
