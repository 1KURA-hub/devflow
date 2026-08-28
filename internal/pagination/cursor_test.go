package pagination

import (
	"errors"
	"testing"
	"time"
)

func TestChronologicalCursorRoundTrip(t *testing.T) {
	wantTime := time.Date(2026, time.August, 4, 10, 11, 12, 345678000, time.UTC)
	raw, err := Encode(Chronological(wantTime, 42))
	if err != nil {
		t.Fatalf("encode cursor: %v", err)
	}

	got, err := Decode(raw, KindChronological)
	if err != nil {
		t.Fatalf("decode cursor: %v", err)
	}
	if got.ID != 42 || !got.CreatedAt().Equal(wantTime) {
		t.Fatalf("round trip mismatch: got id=%d time=%s", got.ID, got.CreatedAt())
	}
}

func TestHotCursorRoundTrip(t *testing.T) {
	raw, err := Encode(Hot(17, 99))
	if err != nil {
		t.Fatalf("encode cursor: %v", err)
	}

	got, err := Decode(raw, KindHot)
	if err != nil {
		t.Fatalf("decode cursor: %v", err)
	}
	if got.Score != 17 || got.ID != 99 {
		t.Fatalf("round trip mismatch: got score=%d id=%d", got.Score, got.ID)
	}
}

func TestCursorKindMismatchIsRejected(t *testing.T) {
	raw, err := Encode(Hot(17, 99))
	if err != nil {
		t.Fatalf("encode cursor: %v", err)
	}
	if _, err := Decode(raw, KindChronological); !errors.Is(err, ErrInvalidCursor) {
		t.Fatalf("expected ErrInvalidCursor, got %v", err)
	}
}

func TestLegacyTimestampCursor(t *testing.T) {
	want := time.Date(2026, time.August, 4, 10, 11, 12, 0, time.FixedZone("CST", 8*60*60))
	got, err := Decode(want.Format(time.RFC3339Nano), KindChronological)
	if err != nil {
		t.Fatalf("decode legacy cursor: %v", err)
	}
	if got.ID != 0 || !got.CreatedAt().Equal(want) {
		t.Fatalf("legacy cursor mismatch: id=%d time=%s", got.ID, got.CreatedAt())
	}
}
