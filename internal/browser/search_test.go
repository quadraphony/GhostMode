package browser

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestDuckDuckGoSearchParsesResults(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`
			<html><body>
				<a class="result__a" href="https://example.com/one">First Result</a>
				<a class="result__a" href="https://example.com/two">Second Result</a>
			</body></html>
		`))
	}))
	defer server.Close()

	results, err := NewDuckDuckGoSearch(server.URL).Search(context.Background(), "ghost mode")
	if err != nil {
		t.Fatalf("Search: %v", err)
	}
	if len(results) != 2 || results[0].URL != "https://example.com/one" {
		t.Fatalf("unexpected results: %+v", results)
	}
}
