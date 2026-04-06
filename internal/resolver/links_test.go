package resolver

import (
	"strings"
	"testing"

	"ghost-browser/pkg/types"
	"golang.org/x/net/html"
)

func TestResolveReference(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		baseURL string
		href    string
		want    string
		ok      bool
	}{
		{"relative", "https://example.com/docs/page", "../about", "https://example.com/about", true},
		{"absolute", "https://example.com/docs/page", "https://golang.org", "https://golang.org", true},
		{"fragment", "https://example.com/docs/page", "#intro", "https://example.com/docs/page#intro", true},
		{"javascript", "https://example.com", "javascript:alert(1)", "", false},
		{"malformed", "https://example.com", "http://[::1", "", false},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got, ok := ResolveReference(tc.baseURL, tc.href)
			if ok != tc.ok {
				t.Fatalf("ResolveReference ok = %v, want %v", ok, tc.ok)
			}
			if got != tc.want {
				t.Fatalf("ResolveReference = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestExtractLinks(t *testing.T) {
	t.Parallel()

	doc, err := html.Parse(strings.NewReader(`
		<html><body>
			<a href="/one"> One </a>
			<a href="https://example.com/two">Two</a>
			<a href="/one">One</a>
			<a href="javascript:void(0)">Bad</a>
			<a href="#frag">Fragment</a>
			<a>No href</a>
		</body></html>
	`))
	if err != nil {
		t.Fatalf("html.Parse: %v", err)
	}

	links := ExtractLinks(doc, "https://example.com/base/page")
	if len(links) != 3 {
		t.Fatalf("len(links) = %d, want 3", len(links))
	}
	if links[0].Index != 1 || links[0].URL != "https://example.com/one" || links[0].Label != "One" {
		t.Fatalf("unexpected first link: %+v", links[0])
	}
	if links[2].URL != "https://example.com/base/page#frag" {
		t.Fatalf("unexpected fragment link: %+v", links[2])
	}
}

func TestSplitLinks(t *testing.T) {
	t.Parallel()

	articles, utility := SplitLinks([]types.Link{
		{Index: 1, Category: types.LinkCategoryArticle},
		{Index: 2, Category: types.LinkCategoryUtility},
		{Index: 3, Category: types.LinkCategoryArticle},
	})
	if len(articles) != 2 || len(utility) != 1 {
		t.Fatalf("unexpected split: articles=%d utility=%d", len(articles), len(utility))
	}
}
