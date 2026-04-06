package types

import "time"

type HistoryEntry struct {
	Title     string    `json:"title"`
	URL       string    `json:"url"`
	VisitedAt time.Time `json:"visited_at"`
}

type Bookmark struct {
	Title     string    `json:"title"`
	URL       string    `json:"url"`
	CreatedAt time.Time `json:"created_at"`
}

type NavigationState struct {
	CurrentPage    *Page
	BackStack      []*Page
	ForwardStack   []*Page
	HistoryEntries []HistoryEntry
}
