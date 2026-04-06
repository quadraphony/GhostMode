package fetcher

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	apperrors "ghost-browser/internal/errors"
	"ghost-browser/pkg/types"
)

const (
	DefaultTimeout   = 10 * time.Second
	DefaultUserAgent = "GhostModeBrowser/0.1 (+https://ghostmode.local)"
	MaxRedirects     = 10
)

type Fetcher struct {
	client    *http.Client
	userAgent string
}

func New(timeout time.Duration, userAgent string) *Fetcher {
	if timeout <= 0 {
		timeout = DefaultTimeout
	}
	if strings.TrimSpace(userAgent) == "" {
		userAgent = DefaultUserAgent
	}

	return &Fetcher{
		client: &http.Client{
			Timeout: timeout,
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				if len(via) >= MaxRedirects {
					return apperrors.ErrTooManyRedirects
				}
				return nil
			},
		},
		userAgent: userAgent,
	}
}

func (f *Fetcher) Fetch(ctx context.Context, targetURL string) (*types.FetchResult, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, targetURL, nil)
	if err != nil {
		return nil, fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("User-Agent", f.userAgent)
	req.Header.Set("Accept", "text/html,application/xhtml+xml")

	resp, err := f.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetch %q: %w", targetURL, err)
	}
	defer resp.Body.Close()

	contentType := resp.Header.Get("Content-Type")
	if !isHTMLContentType(contentType) {
		return nil, fmt.Errorf("%w: %s", apperrors.ErrUnsupportedContent, contentType)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response body: %w", err)
	}
	if len(bytes.TrimSpace(body)) == 0 {
		return nil, fmt.Errorf("%w", apperrors.ErrEmptyResponseBody)
	}

	finalURL := targetURL
	if resp.Request != nil && resp.Request.URL != nil {
		finalURL = resp.Request.URL.String()
	}

	return &types.FetchResult{
		StatusCode:  resp.StatusCode,
		ContentType: contentType,
		Body:        body,
		FinalURL:    finalURL,
		Headers:     resp.Header.Clone(),
	}, nil
}

func isHTMLContentType(value string) bool {
	base := strings.ToLower(strings.TrimSpace(strings.Split(value, ";")[0]))
	return base == "text/html" || base == "application/xhtml+xml"
}
