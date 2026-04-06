package parser

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"ghost-browser/pkg/types"
)

func TestFixtureExtractionQuality(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name              string
		fixture           string
		title             string
		mustContain       []string
		mustNotContain    []string
		minArticleLinks   int
		maxUtilityLinks   int
		expectJSHeavyFlag bool
	}{
		{
			name:            "news article",
			fixture:         "article.html",
			title:           "Ghost Mode Test Page",
			mustContain:     []string{"Ghost Mode Browser keeps the terminal view readable."},
			mustNotContain:  []string{"Home Docs Pricing", "console.log"},
			minArticleLinks: 0,
			maxUtilityLinks: 0,
		},
		{
			name:            "blog article",
			fixture:         "blog_article.html",
			title:           "Engineering Notes",
			mustContain:     []string{"Building trustworthy terminal readers", "Longer paragraphs, coherent flow, and low link density are strong signals"},
			mustNotContain:  []string{"Home Archive About Newsletter", "Share Post Twitter Facebook"},
			minArticleLinks: 0,
			maxUtilityLinks: 0,
		},
		{
			name:            "docs page",
			fixture:         "docs_page.html",
			title:           "API Reference",
			mustContain:     []string{"Use bearer tokens for authenticated requests to the API."},
			mustNotContain:  []string{"Table of contents"},
			minArticleLinks: 0,
			maxUtilityLinks: 0,
		},
		{
			name:            "wiki like page",
			fixture:         "wiki_like.html",
			title:           "Terminal browser - Wiki",
			mustContain:     []string{"A terminal browser is a text-based client used to access web content in command-line environments."},
			mustNotContain:  []string{"Current events", "Privacy policy"},
			minArticleLinks: 0,
			maxUtilityLinks: 0,
		},
		{
			name:            "cookie clutter",
			fixture:         "cookie_clutter.html",
			title:           "Cookie Banner Story",
			mustContain:     []string{"The article body should survive even when aggressive cookie and popup clutter surrounds it."},
			mustNotContain:  []string{"We use cookies", "Subscribe to our newsletter"},
			minArticleLinks: 0,
			maxUtilityLinks: 0,
		},
		{
			name:            "product page",
			fixture:         "product_page.html",
			title:           "Noise-Cancelling Headphones",
			mustContain:     []string{"Wireless over-ear headphones with adaptive noise cancellation"},
			mustNotContain:  []string{"Shop Deals Support Account Cart", "Related products accessories bundles compare"},
			minArticleLinks: 1,
			maxUtilityLinks: 0,
		},
		{
			name:              "js shell",
			fixture:           "js_shell.html",
			title:             "News App",
			mustNotContain:    []string{"__INITIAL_STATE__"},
			minArticleLinks:   0,
			maxUtilityLinks:   0,
			expectJSHeavyFlag: true,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			page := parseFixture(t, tc.fixture)
			if page.Title != tc.title {
				t.Fatalf("Title = %q, want %q", page.Title, tc.title)
			}
			for _, want := range tc.mustContain {
				if !strings.Contains(page.TextContent, want) {
					t.Fatalf("TextContent missing %q: %q", want, page.TextContent)
				}
			}
			for _, unwanted := range tc.mustNotContain {
				if strings.Contains(page.TextContent, unwanted) {
					t.Fatalf("TextContent should not include %q: %q", unwanted, page.TextContent)
				}
			}
			if len(page.ArticleLinks) < tc.minArticleLinks {
				t.Fatalf("len(ArticleLinks) = %d, want at least %d", len(page.ArticleLinks), tc.minArticleLinks)
			}
			if len(page.UtilityLinks) > tc.maxUtilityLinks {
				t.Fatalf("len(UtilityLinks) = %d, want at most %d", len(page.UtilityLinks), tc.maxUtilityLinks)
			}
			if tc.expectJSHeavyFlag && page.Metadata["js_heavy_shell"] != "true" {
				t.Fatalf("expected js_heavy_shell metadata, got %+v", page.Metadata)
			}
		})
	}
}

func TestFixtureReadabilityQuality(t *testing.T) {
	t.Parallel()

	page := parseFixture(t, "sidebar_heavy.html")
	if !strings.Contains(page.ReadabilityContent, "Focused article beats dense chrome") {
		t.Fatalf("unexpected readability content: %q", page.ReadabilityContent)
	}
	if strings.Contains(page.ReadabilityContent, "Sponsored links") {
		t.Fatalf("sidebar text leaked into readability content: %q", page.ReadabilityContent)
	}
}

func parseFixture(t *testing.T, name string) *types.Page {
	t.Helper()

	path := filepath.Join("..", "..", "test", "fixtures", name)
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile(%s): %v", name, err)
	}
	page, err := New().Parse("https://example.com/source", &types.FetchResult{
		ContentType: "text/html; charset=utf-8",
		Body:        data,
		FinalURL:    "https://example.com/" + name,
	})
	if err != nil {
		t.Fatalf("Parse(%s): %v", name, err)
	}
	return page
}
