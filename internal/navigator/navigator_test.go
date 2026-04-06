package navigator

import (
	"testing"
	"time"

	"ghost-browser/pkg/types"
)

func TestNavigatorBackForward(t *testing.T) {
	t.Parallel()

	nav := New(nil)
	nav.Push(&types.Page{FinalURL: "https://example.com/1"}, types.HistoryEntry{URL: "1", VisitedAt: time.Now()})
	nav.Push(&types.Page{FinalURL: "https://example.com/2"}, types.HistoryEntry{URL: "2", VisitedAt: time.Now()})

	back, err := nav.Back()
	if err != nil {
		t.Fatalf("Back: %v", err)
	}
	if back.FinalURL != "https://example.com/1" {
		t.Fatalf("Back page = %q", back.FinalURL)
	}

	forward, err := nav.Forward()
	if err != nil {
		t.Fatalf("Forward: %v", err)
	}
	if forward.FinalURL != "https://example.com/2" {
		t.Fatalf("Forward page = %q", forward.FinalURL)
	}
}
