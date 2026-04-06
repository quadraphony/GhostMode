package fetcher

import (
	"context"
	stderrors "errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	apperrors "ghost-browser/internal/errors"
)

func TestFetchHTMLPage(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("User-Agent"); got != DefaultUserAgent {
			t.Fatalf("User-Agent = %q, want %q", got, DefaultUserAgent)
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("<html><body>ok</body></html>"))
	}))
	defer server.Close()

	result, err := New(DefaultTimeout, DefaultUserAgent).Fetch(context.Background(), server.URL)
	if err != nil {
		t.Fatalf("Fetch returned error: %v", err)
	}

	if result.StatusCode != http.StatusOK {
		t.Fatalf("StatusCode = %d, want %d", result.StatusCode, http.StatusOK)
	}
	if result.FinalURL != server.URL {
		t.Fatalf("FinalURL = %q, want %q", result.FinalURL, server.URL)
	}
	if len(result.Body) == 0 {
		t.Fatal("expected non-empty body")
	}
}

func TestFetchFollowsRedirects(t *testing.T) {
	t.Parallel()

	target := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		_, _ = w.Write([]byte("<html><body>redirect target</body></html>"))
	}))
	defer target.Close()

	redirector := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, target.URL+"/article", http.StatusFound)
	}))
	defer redirector.Close()

	result, err := New(DefaultTimeout, DefaultUserAgent).Fetch(context.Background(), redirector.URL)
	if err != nil {
		t.Fatalf("Fetch returned error: %v", err)
	}
	if result.FinalURL != target.URL+"/article" {
		t.Fatalf("FinalURL = %q, want %q", result.FinalURL, target.URL+"/article")
	}
}

func TestFetchRejectsUnsupportedContentType(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	defer server.Close()

	_, err := New(DefaultTimeout, DefaultUserAgent).Fetch(context.Background(), server.URL)
	if !stderrors.Is(err, apperrors.ErrUnsupportedContent) {
		t.Fatalf("expected ErrUnsupportedContent, got %v", err)
	}
}

func TestFetchTimeout(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(200 * time.Millisecond)
		w.Header().Set("Content-Type", "text/html")
		_, _ = w.Write([]byte("<html><body>slow</body></html>"))
	}))
	defer server.Close()

	_, err := New(50*time.Millisecond, DefaultUserAgent).Fetch(context.Background(), server.URL)
	if err == nil {
		t.Fatal("expected timeout error, got nil")
	}
}

func TestFetchRejectsEmptyHTMLBody(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		_, _ = w.Write([]byte("   "))
	}))
	defer server.Close()

	_, err := New(DefaultTimeout, DefaultUserAgent).Fetch(context.Background(), server.URL)
	if !stderrors.Is(err, apperrors.ErrEmptyResponseBody) {
		t.Fatalf("expected ErrEmptyResponseBody, got %v", err)
	}
}
