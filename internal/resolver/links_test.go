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

	links := ExtractLinks(doc, "https://example.com/base/page", nil)
	if len(links) != 3 {
		t.Fatalf("len(links) = %d, want 3", len(links))
	}
	want := map[string]bool{
		"https://example.com/one":            false,
		"https://example.com/two":            false,
		"https://example.com/base/page#frag": false,
	}
	for _, link := range links {
		if _, ok := want[link.URL]; ok {
			want[link.URL] = true
		}
	}
	for target, seen := range want {
		if !seen {
			t.Fatalf("expected link %q in %+v", target, links)
		}
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

func TestExtractLinksBoostsMainContentLinks(t *testing.T) {
	t.Parallel()

	doc, err := html.Parse(strings.NewReader(`
		<html><body>
			<nav><a href="/home">Home</a></nav>
			<main>
				<article id="article-content">
					<a href="/story">A descriptive article headline with enough detail</a>
				</article>
			</main>
			<footer><a href="/privacy">Privacy</a></footer>
		</body></html>
	`))
	if err != nil {
		t.Fatalf("html.Parse: %v", err)
	}

	main := findByID(doc, "article-content")
	links := ExtractLinks(doc, "https://example.com", main)
	if len(links) != 3 {
		t.Fatalf("len(links) = %d, want 3", len(links))
	}
	if links[0].Category != types.LinkCategoryArticle || !strings.Contains(links[0].URL, "/story") {
		t.Fatalf("expected main content article first, got %+v", links[0])
	}
	if links[2].Category != types.LinkCategoryUtility {
		t.Fatalf("expected utility link to be demoted, got %+v", links[2])
	}
}

func TestExtractLinksDedupesByURL(t *testing.T) {
	t.Parallel()

	doc, err := html.Parse(strings.NewReader(`
		<html><body>
			<main id="content">
				<a href="/story">Short</a>
				<a href="/story">A much more descriptive article title</a>
			</main>
		</body></html>
	`))
	if err != nil {
		t.Fatalf("html.Parse: %v", err)
	}

	links := ExtractLinks(doc, "https://example.com", findByID(doc, "content"))
	if len(links) != 1 {
		t.Fatalf("len(links) = %d, want 1", len(links))
	}
	if links[0].Label != "A much more descriptive article title" {
		t.Fatalf("unexpected deduped link: %+v", links[0])
	}
}

func findByID(node *html.Node, want string) *html.Node {
	if node == nil {
		return nil
	}
	if node.Type == html.ElementNode {
		for _, attr := range node.Attr {
			if attr.Key == "id" && attr.Val == want {
				return node
			}
		}
	}
	for child := node.FirstChild; child != nil; child = child.NextSibling {
		if found := findByID(child, want); found != nil {
			return found
		}
	}
	return nil
}
