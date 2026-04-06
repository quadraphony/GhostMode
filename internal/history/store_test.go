package history

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"ghost-browser/pkg/types"
)

func TestStoreRoundTrip(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	store := NewStore(dir)
	input := []types.HistoryEntry{{Title: "Example", URL: "https://example.com", VisitedAt: time.Unix(10, 0).UTC()}}
	if err := store.Save(input); err != nil {
		t.Fatalf("Save: %v", err)
	}

	got, err := store.Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if len(got) != 1 || got[0].URL != input[0].URL {
		t.Fatalf("unexpected history: %+v", got)
	}
}

func TestStoreRejectsCorruptJSON(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, historyFile), []byte("{"), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}
	_, err := NewStore(dir).Load()
	if err == nil {
		t.Fatal("expected error for corrupt JSON")
	}
}
