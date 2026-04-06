package parser

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"ghost-browser/pkg/types"
)

func TestParseExtractsReadableContent(t *testing.T) {
	t.Parallel()

	body := readFixture(t, "article.html")
	page, err := New().Parse("https://example.com/source", &types.FetchResult{
		ContentType: "text/html; charset=utf-8",
		Body:        []byte(body),
		FinalURL:    "https://example.com/article",
	})
	if err != nil {
		t.Fatalf("Parse returned error: %v", err)
	}

	if page.Title != "Ghost Mode Test Page" {
		t.Fatalf("Title = %q, want %q", page.Title, "Ghost Mode Test Page")
	}
	if strings.Contains(page.TextContent, "console.log") {
		t.Fatalf("TextContent should not include script content: %q", page.TextContent)
	}
	if strings.Contains(page.TextContent, ".hero") {
		t.Fatalf("TextContent should not include style content: %q", page.TextContent)
	}
	if strings.Contains(page.TextContent, "Hidden helper copy") {
		t.Fatalf("TextContent should not include hidden content: %q", page.TextContent)
	}
	if !strings.Contains(page.TextContent, "Ghost Mode Browser keeps the terminal view readable.") {
		t.Fatalf("TextContent missing article text: %q", page.TextContent)
	}
	if !strings.Contains(page.TextContent, "Second paragraph with extra spacing.") {
		t.Fatalf("TextContent missing normalized paragraph: %q", page.TextContent)
	}
	if strings.Contains(page.TextContent, "   ") {
		t.Fatalf("TextContent should normalize whitespace: %q", page.TextContent)
	}
}

func TestParseHandlesMalformedHTML(t *testing.T) {
	t.Parallel()

	body := readFixture(t, "malformed.html")
	page, err := New().Parse("https://example.com/source", &types.FetchResult{
		ContentType: "text/html",
		Body:        []byte(body),
		FinalURL:    "https://example.com/broken",
	})
	if err != nil {
		t.Fatalf("Parse returned error for malformed HTML: %v", err)
	}

	if page.Title != "Broken Title" {
		t.Fatalf("Title = %q, want %q", page.Title, "Broken Title")
	}
	if !strings.Contains(page.TextContent, "Broken but readable paragraph") {
		t.Fatalf("expected extracted text, got %q", page.TextContent)
	}
}

func TestParseWarnsWhenNoReadableText(t *testing.T) {
	t.Parallel()

	page, err := New().Parse("https://example.com/source", &types.FetchResult{
		ContentType: "text/html",
		Body:        []byte("<html><head><title>Only Script</title></head><body><script>ignored()</script></body></html>"),
		FinalURL:    "https://example.com/empty",
	})
	if err != nil {
		t.Fatalf("Parse returned error: %v", err)
	}

	if page.TextContent != "" {
		t.Fatalf("TextContent = %q, want empty", page.TextContent)
	}
	if len(page.Warnings) != 1 {
		t.Fatalf("Warnings length = %d, want 1", len(page.Warnings))
	}
}

func TestParseFiltersJunkContainersByClassAndID(t *testing.T) {
	t.Parallel()

	page, err := New().Parse("https://example.com/source", &types.FetchResult{
		ContentType: "text/html",
		Body: []byte(`
			<html><body>
				<div id="site-header">Top nav and sign in</div>
				<div class="cookie-banner">We use cookies and share analytics</div>
				<main>
					<article class="story-body">
						<p>Real article text stays visible.</p>
					</article>
				</main>
				<aside class="related-links">related stories</aside>
			</body></html>
		`),
		FinalURL: "https://example.com/noisy",
	})
	if err != nil {
		t.Fatalf("Parse returned error: %v", err)
	}

	for _, unwanted := range []string{"Top nav and sign in", "We use cookies", "related stories"} {
		if strings.Contains(page.TextContent, unwanted) {
			t.Fatalf("TextContent should not include %q: %q", unwanted, page.TextContent)
		}
	}
	if !strings.Contains(page.TextContent, "Real article text stays visible.") {
		t.Fatalf("expected article content to remain, got %q", page.TextContent)
	}
}

func TestParsePrefersPrimaryContentSubtree(t *testing.T) {
	t.Parallel()

	body := readFixture(t, "noisy_layout.html")
	page, err := New().Parse("https://example.com/source", &types.FetchResult{
		ContentType: "text/html",
		Body:        []byte(body),
		FinalURL:    "https://example.com/noisy-layout",
	})
	if err != nil {
		t.Fatalf("Parse returned error: %v", err)
	}

	for _, unwanted := range []string{"Subscribe now for premium access", "Privacy Terms Contact Careers", "Home News Business"} {
		if strings.Contains(page.TextContent, unwanted) {
			t.Fatalf("TextContent should not include %q: %q", unwanted, page.TextContent)
		}
	}
	for _, wanted := range []string{
		"GhostMode should extract the primary article instead of dragging the entire page into the terminal.",
		"When the parser prefers the main story container first, the output becomes much easier to trust.",
	} {
		if !strings.Contains(page.TextContent, wanted) {
			t.Fatalf("TextContent missing %q: %q", wanted, page.TextContent)
		}
	}
}

func TestParseFallsBackWhenNoStrongCandidateExists(t *testing.T) {
	t.Parallel()

	page, err := New().Parse("https://example.com/source", &types.FetchResult{
		ContentType: "text/html",
		Body: []byte(`
			<html><body>
				<section><p>Short note one.</p></section>
				<section><p>Short note two.</p></section>
			</body></html>
		`),
		FinalURL: "https://example.com/fallback",
	})
	if err != nil {
		t.Fatalf("Parse returned error: %v", err)
	}

	if !strings.Contains(page.TextContent, "Short note one.") || !strings.Contains(page.TextContent, "Short note two.") {
		t.Fatalf("expected fallback extraction to keep body content, got %q", page.TextContent)
	}
}

func readFixture(t *testing.T, name string) string {
	t.Helper()

	path := filepath.Join("..", "..", "test", "fixtures", name)
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read fixture %s: %v", name, err)
	}

	return string(data)
}
