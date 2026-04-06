package app

import (
	"context"
	stderrors "errors"
	"fmt"
	"io"
	"net"
	"net/url"
	"strings"

	apperrors "ghost-browser/internal/errors"
	"ghost-browser/internal/fetcher"
	"ghost-browser/internal/resolver"
)

func Run(args []string, stdout, stderr io.Writer) int {
	if len(args) == 0 {
		fmt.Fprintln(stderr, "usage: ghost <url>")
		return 1
	}

	normalizedURL, err := resolver.NormalizeURL(args[0])
	if err != nil {
		fmt.Fprintf(stderr, "error: %s\n", formatError(err))
		return 1
	}

	result, err := fetcher.New(fetcher.DefaultTimeout, fetcher.DefaultUserAgent).Fetch(context.Background(), normalizedURL)
	if err != nil {
		fmt.Fprintf(stderr, "error: %s\n", formatError(err))
		return 1
	}

	fmt.Fprintln(stdout, "Ghost Mode Browser")
	fmt.Fprintf(stdout, "Source URL: %s\n", normalizedURL)
	fmt.Fprintf(stdout, "Final URL: %s\n", result.FinalURL)
	fmt.Fprintf(stdout, "Status: %d\n", result.StatusCode)
	fmt.Fprintf(stdout, "Content-Type: %s\n", result.ContentType)
	fmt.Fprintf(stdout, "Bytes: %d\n", len(result.Body))

	return 0
}

func formatError(err error) string {
	switch {
	case stderrors.Is(err, apperrors.ErrEmptyURL):
		return "URL is required"
	case stderrors.Is(err, apperrors.ErrInvalidURL):
		return err.Error()
	case stderrors.Is(err, apperrors.ErrUnsupportedScheme):
		return err.Error()
	case stderrors.Is(err, apperrors.ErrUnsupportedContent):
		return err.Error()
	case stderrors.Is(err, apperrors.ErrTooManyRedirects):
		return "request failed after too many redirects"
	case stderrors.Is(err, apperrors.ErrEmptyResponseBody):
		return "received an empty HTML response body"
	}

	var dnsErr *net.DNSError
	if stderrors.As(err, &dnsErr) {
		return fmt.Sprintf("DNS lookup failed for %q", dnsErr.Name)
	}

	var urlErr *url.Error
	if stderrors.As(err, &urlErr) {
		if urlErr.Timeout() {
			return fmt.Sprintf("request timed out for %s", urlErr.URL)
		}
		return strings.TrimSpace(urlErr.Err.Error())
	}

	return err.Error()
}
