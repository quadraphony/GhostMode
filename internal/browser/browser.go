package browser

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"

	"ghost-browser/internal/bookmarks"
	apperrors "ghost-browser/internal/errors"
	"ghost-browser/internal/fetcher"
	"ghost-browser/internal/history"
	"ghost-browser/internal/navigator"
	"ghost-browser/internal/parser"
	"ghost-browser/internal/renderer"
	"ghost-browser/internal/resolver"
	"ghost-browser/pkg/types"
)

type SearchProvider interface {
	Search(ctx context.Context, query string) ([]types.SearchResult, error)
}

type Browser struct {
	fetcher        *fetcher.Fetcher
	parser         *parser.Parser
	rendererWidth  int
	nav            *navigator.Navigator
	bookmarks      *bookmarks.Store
	history        *history.Store
	searchProvider SearchProvider
	readability    bool
}

func New(fetch *fetcher.Fetcher, parse *parser.Parser, bookmarkStore *bookmarks.Store, historyStore *history.Store, search SearchProvider, historyEntries []types.HistoryEntry) *Browser {
	return &Browser{
		fetcher:        fetch,
		parser:         parse,
		nav:            navigator.New(historyEntries),
		bookmarks:      bookmarkStore,
		history:        historyStore,
		searchProvider: search,
		rendererWidth:  88,
	}
}

func (b *Browser) LoadURL(ctx context.Context, source string) (*types.Page, error) {
	normalized, err := resolver.NormalizeURL(source)
	if err != nil {
		return nil, err
	}
	result, err := b.fetcher.Fetch(ctx, normalized)
	if err != nil {
		return nil, err
	}
	page, err := b.parser.Parse(normalized, result)
	if err != nil {
		return nil, err
	}
	b.nav.Push(page, types.HistoryEntry{
		Title:     page.Title,
		URL:       page.FinalURL,
		VisitedAt: time.Now().UTC(),
	})
	if err := b.history.Save(b.nav.History()); err != nil {
		page.Warnings = append(page.Warnings, fmt.Sprintf("History save failed: %v", err))
	}
	return page, nil
}

func (b *Browser) Reload(ctx context.Context) (*types.Page, error) {
	current := b.nav.Current()
	if current == nil {
		return nil, navigator.ErrNoCurrentPage
	}
	result, err := b.fetcher.Fetch(ctx, current.FinalURL)
	if err != nil {
		return nil, err
	}
	page, err := b.parser.Parse(current.SourceURL, result)
	if err != nil {
		return nil, err
	}
	b.nav.ReplaceCurrent(page)
	return page, nil
}

func (b *Browser) Back() (*types.Page, error) {
	return b.nav.Back()
}

func (b *Browser) Forward() (*types.Page, error) {
	return b.nav.Forward()
}

func (b *Browser) RenderCurrent() string {
	return renderer.Render(b.nav.Current(), renderer.Options{
		Width:           b.rendererWidth,
		ReadabilityMode: b.readability,
		ShowHelpHint:    true,
	})
}

func (b *Browser) ToggleReadability() string {
	b.readability = !b.readability
	if b.readability {
		return "Readability mode enabled."
	}
	return "Readability mode disabled."
}

func (b *Browser) OpenLink(ctx context.Context, index int) (*types.Page, error) {
	current := b.nav.Current()
	if current == nil {
		return nil, navigator.ErrNoCurrentPage
	}
	for _, link := range current.Links {
		if link.Index == index {
			return b.LoadURL(ctx, link.URL)
		}
	}
	return nil, fmt.Errorf("link %d not found", index)
}

func (b *Browser) AddBookmark() (string, error) {
	current := b.nav.Current()
	if current == nil {
		return "", navigator.ErrNoCurrentPage
	}
	existing, err := b.bookmarks.Load()
	if err != nil {
		return "", err
	}
	for _, entry := range existing {
		if entry.URL == current.FinalURL {
			return "Bookmark already exists.", nil
		}
	}
	existing = append(existing, types.Bookmark{
		Title:     fallbackTitle(current),
		URL:       current.FinalURL,
		CreatedAt: time.Now().UTC(),
	})
	if err := b.bookmarks.Save(existing); err != nil {
		return "", err
	}
	return "Bookmark added.", nil
}

func (b *Browser) ListBookmarks() ([]types.Bookmark, error) {
	return b.bookmarks.Load()
}

func (b *Browser) OpenBookmark(ctx context.Context, index int) (*types.Page, error) {
	items, err := b.bookmarks.Load()
	if err != nil {
		return nil, err
	}
	if index < 1 || index > len(items) {
		return nil, fmt.Errorf("bookmark %d not found", index)
	}
	return b.LoadURL(ctx, items[index-1].URL)
}

func (b *Browser) HistoryEntries() []types.HistoryEntry {
	return b.nav.History()
}

func (b *Browser) Search(ctx context.Context, query string) (*types.Page, error) {
	if b.searchProvider == nil {
		return nil, errors.New("search is not configured")
	}
	results, err := b.searchProvider.Search(ctx, query)
	if err != nil {
		return nil, err
	}
	page := &types.Page{
		SourceURL: "search:" + query,
		FinalURL:  "search:" + query,
		Title:     "Search: " + query,
		Metadata:  map[string]string{"search_query": query},
	}
	var lines []string
	for i, result := range results {
		page.Links = append(page.Links, types.Link{Index: i + 1, Label: result.Title, URL: result.URL})
		line := result.Title
		if result.Snippet != "" {
			line += "\n\n" + result.Snippet
		}
		lines = append(lines, line)
	}
	page.TextContent = strings.Join(lines, "\n\n")
	page.ReadabilityContent = page.TextContent
	b.nav.Push(page, types.HistoryEntry{
		Title:     page.Title,
		URL:       page.FinalURL,
		VisitedAt: time.Now().UTC(),
	})
	if err := b.history.Save(b.nav.History()); err != nil {
		page.Warnings = append(page.Warnings, fmt.Sprintf("History save failed: %v", err))
	}
	return page, nil
}

func (b *Browser) RunInteractive(ctx context.Context, in io.Reader, out, errOut io.Writer) error {
	reader := bufio.NewScanner(in)
	for {
		fmt.Fprint(out, "\nghost> ")
		if !reader.Scan() {
			return nil
		}
		line := strings.TrimSpace(reader.Text())
		if line == "" {
			continue
		}

		message, shouldQuit, err := b.Execute(ctx, line, out)
		if err != nil {
			fmt.Fprintf(errOut, "error: %s\n", formatError(err))
			continue
		}
		if message != "" {
			fmt.Fprintln(out, message)
		}
		if shouldQuit {
			return nil
		}
	}
}

func (b *Browser) Execute(ctx context.Context, command string, out io.Writer) (string, bool, error) {
	cmd, arg := parseCommand(command)
	switch cmd {
	case "quit", "exit":
		return "Bye.", true, nil
	case "help":
		return helpText(), false, nil
	case "reload":
		_, err := b.Reload(ctx)
		if err != nil {
			return "", false, err
		}
		fmt.Fprint(out, b.RenderCurrent())
		return "", false, nil
	case "back":
		_, err := b.Back()
		if err != nil {
			return "", false, err
		}
		fmt.Fprint(out, b.RenderCurrent())
		return "", false, nil
	case "forward":
		_, err := b.Forward()
		if err != nil {
			return "", false, err
		}
		fmt.Fprint(out, b.RenderCurrent())
		return "", false, nil
	case "open":
		return b.execOpen(ctx, arg, out)
	case "bookmark":
		return b.execBookmark(ctx, arg, out)
	case "history":
		return renderHistory(b.HistoryEntries()), false, nil
	case "search":
		if strings.TrimSpace(arg) == "" {
			return "", false, errors.New("search query is required")
		}
		_, err := b.Search(ctx, arg)
		if err != nil {
			return "", false, err
		}
		fmt.Fprint(out, b.RenderCurrent())
		return "", false, nil
	case "readability":
		return b.ToggleReadability(), false, nil
	default:
		if _, ok := resolver.ResolveReference("https://example.com", command); ok {
			_, err := b.LoadURL(ctx, command)
			if err != nil {
				return "", false, err
			}
			fmt.Fprint(out, b.RenderCurrent())
			return "", false, nil
		}
		return "", false, fmt.Errorf("unknown command: %s", command)
	}
}

func (b *Browser) execOpen(ctx context.Context, arg string, out io.Writer) (string, bool, error) {
	arg = strings.TrimSpace(arg)
	if arg == "" {
		return "", false, errors.New("open requires a link number, bookmark reference, or URL")
	}
	if strings.HasPrefix(arg, "bookmark ") {
		index, err := parsePositiveInt(strings.TrimSpace(strings.TrimPrefix(arg, "bookmark ")))
		if err != nil {
			return "", false, err
		}
		_, err = b.OpenBookmark(ctx, index)
		if err != nil {
			return "", false, err
		}
		fmt.Fprint(out, b.RenderCurrent())
		return "", false, nil
	}
	if index, err := parsePositiveInt(arg); err == nil {
		_, err = b.OpenLink(ctx, index)
		if err != nil {
			return "", false, err
		}
		fmt.Fprint(out, b.RenderCurrent())
		return "", false, nil
	}
	_, err := b.LoadURL(ctx, arg)
	if err != nil {
		return "", false, err
	}
	fmt.Fprint(out, b.RenderCurrent())
	return "", false, nil
}

func (b *Browser) execBookmark(ctx context.Context, arg string, out io.Writer) (string, bool, error) {
	switch strings.TrimSpace(arg) {
	case "add":
		message, err := b.AddBookmark()
		return message, false, err
	case "list":
		items, err := b.ListBookmarks()
		if err != nil {
			return "", false, err
		}
		return renderBookmarks(items), false, nil
	default:
		return "", false, errors.New("bookmark commands: add, list")
	}
}

func helpText() string {
	return "Commands: open <number|url>, open bookmark <n>, back, forward, reload, bookmark add, bookmark list, history, search <query>, readability, help, quit"
}

func renderBookmarks(items []types.Bookmark) string {
	if len(items) == 0 {
		return "No bookmarks saved."
	}
	var lines []string
	for i, item := range items {
		lines = append(lines, fmt.Sprintf("[%d] %s", i+1, item.Title))
		lines = append(lines, "    "+item.URL)
	}
	return strings.Join(lines, "\n")
}

func renderHistory(items []types.HistoryEntry) string {
	if len(items) == 0 {
		return "No history yet."
	}
	var lines []string
	for i := len(items) - 1; i >= 0; i-- {
		entry := items[i]
		lines = append(lines, fmt.Sprintf("%s  %s", entry.VisitedAt.Format(time.RFC3339), entry.URL))
	}
	return strings.Join(lines, "\n")
}

func parseCommand(input string) (string, string) {
	parts := strings.Fields(strings.TrimSpace(input))
	if len(parts) == 0 {
		return "", ""
	}
	if len(parts) == 1 {
		return strings.ToLower(parts[0]), ""
	}
	return strings.ToLower(parts[0]), strings.TrimSpace(input[len(parts[0]):])
}

func parsePositiveInt(value string) (int, error) {
	var n int
	_, err := fmt.Sscanf(value, "%d", &n)
	if err != nil || n < 1 {
		return 0, errors.New("expected a positive number")
	}
	return n, nil
}

func fallbackTitle(page *types.Page) string {
	if strings.TrimSpace(page.Title) != "" {
		return page.Title
	}
	return page.FinalURL
}

func formatError(err error) string {
	switch {
	case errors.Is(err, apperrors.ErrEmptyURL), errors.Is(err, apperrors.ErrInvalidURL), errors.Is(err, apperrors.ErrUnsupportedScheme),
		errors.Is(err, apperrors.ErrUnsupportedContent), errors.Is(err, apperrors.ErrTooManyRedirects), errors.Is(err, apperrors.ErrEmptyResponseBody), errors.Is(err, apperrors.ErrInvalidHTML):
		return err.Error()
	case errors.Is(err, navigator.ErrNoCurrentPage), errors.Is(err, navigator.ErrNoBackHistory), errors.Is(err, navigator.ErrNoForward):
		return err.Error()
	default:
		return err.Error()
	}
}
