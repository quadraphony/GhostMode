package bookmarks

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"ghost-browser/pkg/types"
)

const bookmarksFile = "bookmarks.json"

type Store struct {
	dir string
}

func NewStore(dir string) *Store {
	return &Store{dir: dir}
}

func DefaultDir() (string, error) {
	base, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("resolve config dir: %w", err)
	}
	return filepath.Join(base, "ghostmode"), nil
}

func (s *Store) Load() ([]types.Bookmark, error) {
	path := filepath.Join(s.dir, bookmarksFile)
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("read bookmarks: %w", err)
	}
	var bookmarks []types.Bookmark
	if err := json.Unmarshal(data, &bookmarks); err != nil {
		return nil, fmt.Errorf("parse bookmarks: %w", err)
	}
	return bookmarks, nil
}

func (s *Store) Save(bookmarks []types.Bookmark) error {
	if err := os.MkdirAll(s.dir, 0o755); err != nil {
		return fmt.Errorf("create bookmark dir: %w", err)
	}
	path := filepath.Join(s.dir, bookmarksFile)
	data, err := json.MarshalIndent(bookmarks, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal bookmarks: %w", err)
	}
	return os.WriteFile(path, data, 0o644)
}
