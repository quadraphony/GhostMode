package browser

import (
	"bytes"
	"context"
	"testing"
	"time"

	"ghost-browser/internal/bookmarks"
	"ghost-browser/internal/fetcher"
	"ghost-browser/internal/history"
	"ghost-browser/internal/parser"
	"ghost-browser/pkg/types"
)

type fakeSearch struct {
	results []types.SearchResult
}

func (f fakeSearch) Search(context.Context, string) ([]types.SearchResult, error) {
	return f.results, nil
}

func TestExecuteParsesCommands(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	b := New(fetcher.New(fetcher.DefaultTimeout, fetcher.DefaultUserAgent), parser.New(), bookmarks.NewStore(dir), history.NewStore(dir), fakeSearch{
		results: []types.SearchResult{{Title: "Go", URL: "https://golang.org", Snippet: "The Go programming language"}},
	}, nil)
	b.nav.Push(&types.Page{
		Title:       "Start",
		FinalURL:    "https://example.com",
		TextContent: "Hello",
		Links: []types.Link{
			{Index: 1, Label: "Next", URL: "https://example.com/next", Category: types.LinkCategoryUtility},
			{Index: 2, Label: "Long article headline for users", URL: "https://example.com/news/story", Category: types.LinkCategoryArticle},
		},
		ArticleLinks: []types.Link{
			{Index: 2, Label: "Long article headline for users", URL: "https://example.com/news/story", Category: types.LinkCategoryArticle},
		},
	}, types.HistoryEntry{Title: "Start", URL: "https://example.com", VisitedAt: time.Now()})

	msg, quit, err := b.Execute(context.Background(), "bookmark add", &bytes.Buffer{})
	if err != nil || quit || msg != "Bookmark added." {
		t.Fatalf("bookmark add => msg=%q quit=%v err=%v", msg, quit, err)
	}

	msg, quit, err = b.Execute(context.Background(), "bookmark list", &bytes.Buffer{})
	if err != nil || quit || msg == "" {
		t.Fatalf("bookmark list => msg=%q quit=%v err=%v", msg, quit, err)
	}

	msg, quit, err = b.Execute(context.Background(), "search go language", &bytes.Buffer{})
	if err != nil || quit || msg != "" {
		t.Fatalf("search => msg=%q quit=%v err=%v", msg, quit, err)
	}

	msg, quit, err = b.Execute(context.Background(), "readability", &bytes.Buffer{})
	if err != nil || quit || msg == "" {
		t.Fatalf("readability => msg=%q quit=%v err=%v", msg, quit, err)
	}

	msg, quit, err = b.Execute(context.Background(), "articles", &bytes.Buffer{})
	if err != nil || quit || msg == "" {
		t.Fatalf("articles => msg=%q quit=%v err=%v", msg, quit, err)
	}
}
