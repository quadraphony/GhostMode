package app

import (
	"context"
	stderrors "errors"
	"fmt"
	"io"
	"net"
	"net/url"
	"strings"

	"ghost-browser/internal/bookmarks"
	"ghost-browser/internal/browser"
	apperrors "ghost-browser/internal/errors"
	"ghost-browser/internal/fetcher"
	"ghost-browser/internal/history"
	"ghost-browser/internal/parser"
)

func Run(args []string, stdin io.Reader, stdout, stderr io.Writer) int {
	configDir, err := bookmarks.DefaultDir()
	if err != nil {
		fmt.Fprintf(stderr, "error: %s\n", formatError(err))
		return 1
	}

	historyStore := history.NewStore(configDir)
	historyEntries, err := historyStore.Load()
	if err != nil {
		fmt.Fprintf(stderr, "warning: %s\n", formatError(err))
	}

	appBrowser := browser.New(
		fetcher.New(fetcher.DefaultTimeout, fetcher.DefaultUserAgent),
		parser.New(),
		bookmarks.NewStore(configDir),
		historyStore,
		browser.NewDuckDuckGoSearch(""),
		historyEntries,
	)

	ctx := context.Background()
	if len(args) > 0 {
		if _, err := appBrowser.LoadURL(ctx, args[0]); err != nil {
			fmt.Fprintf(stderr, "error: %s\n", formatError(err))
			return 1
		}
		fmt.Fprint(stdout, appBrowser.RenderCurrent())
	} else {
		fmt.Fprintln(stdout, "Ghost Mode Browser")
		fmt.Fprintln(stdout, "Type `help` for commands or paste a URL to begin.")
	}

	if err := appBrowser.RunInteractive(ctx, stdin, stdout, stderr); err != nil {
		fmt.Fprintf(stderr, "error: %s\n", formatError(err))
		return 1
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
