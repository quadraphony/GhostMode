package history

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"ghost-browser/pkg/types"
)

const historyFile = "history.json"

type Store struct {
	dir string
}

func NewStore(dir string) *Store {
	return &Store{dir: dir}
}

func (s *Store) Load() ([]types.HistoryEntry, error) {
	path := filepath.Join(s.dir, historyFile)
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("read history: %w", err)
	}
	var entries []types.HistoryEntry
	if err := json.Unmarshal(data, &entries); err != nil {
		return nil, fmt.Errorf("parse history: %w", err)
	}
	return entries, nil
}

func (s *Store) Save(entries []types.HistoryEntry) error {
	if err := os.MkdirAll(s.dir, 0o755); err != nil {
		return fmt.Errorf("create history dir: %w", err)
	}
	path := filepath.Join(s.dir, historyFile)
	data, err := json.MarshalIndent(entries, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal history: %w", err)
	}
	return os.WriteFile(path, data, 0o644)
}
