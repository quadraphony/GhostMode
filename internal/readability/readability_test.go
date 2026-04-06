package readability

import (
	"os"
	"path/filepath"
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

func TestAnalyzePrefersArticleFixture(t *testing.T) {
	t.Parallel()

	result := analyzeFixture(t, "article.html")
	if result.Node == nil || !strings.Contains(result.Text, "Ghost Mode Browser keeps the terminal view readable.") {
		t.Fatalf("unexpected readability result: %+v", result)
	}
}

func TestAnalyzePrefersBlogContent(t *testing.T) {
	t.Parallel()

	result := analyzeFixture(t, "blog_article.html")
	if !strings.Contains(result.Text, "Longer paragraphs, coherent flow, and low link density are strong signals") {
		t.Fatalf("unexpected readability text: %q", result.Text)
	}
	if strings.Contains(result.Text, "Share Post Twitter Facebook") {
		t.Fatalf("did not expect sidebar share tools in readability output: %q", result.Text)
	}
}

func TestAnalyzePrefersDocsContent(t *testing.T) {
	t.Parallel()

	result := analyzeFixture(t, "docs_page.html")
	if !strings.Contains(result.Text, "Tokens expire after one hour") {
		t.Fatalf("unexpected readability text: %q", result.Text)
	}
	if strings.Contains(result.Text, "Table of contents") {
		t.Fatalf("did not expect docs sidebar text in readability output: %q", result.Text)
	}
}

func TestAnalyzeRejectsSidebarHeavyLayout(t *testing.T) {
	t.Parallel()

	result := analyzeFixture(t, "sidebar_heavy.html")
	if !strings.Contains(result.Text, "Focused article beats dense chrome") {
		t.Fatalf("unexpected readability text: %q", result.Text)
	}
	if strings.Contains(result.Text, "Sponsored links") {
		t.Fatalf("sidebar text leaked into readability output: %q", result.Text)
	}
}

func TestAnalyzeHandlesNoisyHeaderFooter(t *testing.T) {
	t.Parallel()

	result := analyzeFixture(t, "noisy_layout.html")
	if !strings.Contains(result.Text, "Terminal readers need cleaner defaults") {
		t.Fatalf("unexpected readability text: %q", result.Text)
	}
	if strings.Contains(result.Text, "Privacy Terms Contact Careers") {
		t.Fatalf("footer text leaked into readability output: %q", result.Text)
	}
}

func TestExtractFallbackWhenNoCandidate(t *testing.T) {
	t.Parallel()

	doc, err := html.Parse(strings.NewReader(`<html><body><div>tiny</div><div>shell</div></body></html>`))
	if err != nil {
		t.Fatalf("html.Parse: %v", err)
	}
	if got := Extract(doc); got != "" {
		t.Fatalf("expected empty readability fallback, got %q", got)
	}
}

func analyzeFixture(t *testing.T, name string) Result {
	t.Helper()

	path := filepath.Join("..", "..", "test", "fixtures", name)
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile(%s): %v", name, err)
	}
	doc, err := html.Parse(strings.NewReader(string(data)))
	if err != nil {
		t.Fatalf("html.Parse(%s): %v", name, err)
	}
	return Analyze(doc)
}
