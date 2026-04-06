package readability

import (
	"strings"
	"testing"

	"golang.org/x/net/html"
)

func TestExtractPrefersMainContent(t *testing.T) {
	t.Parallel()

	doc, err := html.Parse(strings.NewReader(`
		<html><body>
			<nav>home pricing docs support account menu links</nav>
			<div class="sidebar">tiny side content</div>
			<article class="post-content">
				<h1>Title</h1>
				<p>This is a long article paragraph with enough words to win the scoring process and be selected for readability mode.</p>
				<p>This is the second paragraph that should remain visible in readability mode.</p>
			</article>
		</body></html>
	`))
	if err != nil {
		t.Fatalf("html.Parse: %v", err)
	}

	got := Extract(doc)
	if !strings.Contains(got, "second paragraph") {
		t.Fatalf("unexpected readability output: %q", got)
	}
	if strings.Contains(got, "sidebar") {
		t.Fatalf("readability output should exclude sidebar text: %q", got)
	}
}
