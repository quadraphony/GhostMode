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
	"ghost-browser/internal/parser"
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

	page, err := parser.New().Parse(normalizedURL, result)
	if err != nil {
		fmt.Fprintf(stderr, "error: %s\n", formatError(err))
		return 1
	}

	fmt.Fprintln(stdout, "Ghost Mode Browser")
	if page.Title != "" {
		fmt.Fprintf(stdout, "Title: %s\n", page.Title)
	}
	fmt.Fprintf(stdout, "Source URL: %s\n", page.SourceURL)
	fmt.Fprintf(stdout, "Final URL: %s\n", page.FinalURL)
	fmt.Fprintf(stdout, "Status: %d\n", result.StatusCode)
	fmt.Fprintf(stdout, "Content-Type: %s\n\n", result.ContentType)
	if page.TextContent != "" {
		fmt.Fprintln(stdout, page.TextContent)
		fmt.Fprintln(stdout)
	}
	for _, warning := range page.Warnings {
		fmt.Fprintf(stdout, "Warning: %s\n", warning)
	}

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
	case stderrors.Is(err, apperrors.ErrInvalidHTML):
		return err.Error()
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
