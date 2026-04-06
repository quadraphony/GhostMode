package renderer

import (
	"strings"
	"testing"

	"ghost-browser/pkg/types"
)

func TestRenderIncludesLinksAndContent(t *testing.T) {
	t.Parallel()

	out := Render(&types.Page{
		Title:       "Example",
		FinalURL:    "https://example.com",
		TextContent: "This is readable text for the terminal browser.",
		ArticleLinks: []types.Link{
			{Index: 1, Label: "One Article Headline", URL: "https://example.com/one", Category: types.LinkCategoryArticle},
		},
		UtilityLinks: []types.Link{
			{Index: 2, Label: "Home", URL: "https://example.com/home", Category: types.LinkCategoryUtility},
		},
	}, Options{Width: 40, ShowHelpHint: true})

	for _, want := range []string{"Example", "This is readable text", "Articles", "[1] One Article Headline", "Navigation", "Commands: open <n>"} {
		if !strings.Contains(out, want) {
			t.Fatalf("render output missing %q: %q", want, out)
		}
	}
}

func TestRenderUsesReadabilityContent(t *testing.T) {
	t.Parallel()

	out := Render(&types.Page{
		FinalURL:           "https://example.com",
		TextContent:        "plain",
		ReadabilityContent: "clean article",
	}, Options{ReadabilityMode: true})

	if !strings.Contains(out, "clean article") {
		t.Fatalf("expected readability content, got %q", out)
	}
	if strings.Contains(out, "plain") {
		t.Fatalf("did not expect standard content, got %q", out)
	}
}

func TestRenderSuppressesEmptyLinkSections(t *testing.T) {
	t.Parallel()

	out := Render(&types.Page{
		Title:       "Empty",
		FinalURL:    "https://example.com/really/long/url/that/should/be/truncated/for/terminal/output/because/it/is/far/too/long",
		TextContent: "Body text.",
	}, Options{Width: 50})

	if strings.Contains(out, "Navigation\n") || strings.Contains(out, "Articles\n") {
		t.Fatalf("did not expect empty link sections in %q", out)
	}
	if !strings.Contains(out, "...") {
		t.Fatalf("expected truncated URL in %q", out)
	}
}
