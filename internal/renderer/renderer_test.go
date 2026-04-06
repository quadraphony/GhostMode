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
		Links: []types.Link{
			{Index: 1, Label: "One", URL: "https://example.com/one"},
		},
	}, Options{Width: 40, ShowHelpHint: true})

	for _, want := range []string{"Title: Example", "This is readable text", "[1] One", "Commands: open <n>"} {
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
