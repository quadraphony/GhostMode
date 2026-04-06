package resolver

import (
	"fmt"
	"net/url"
	"strings"

	apperrors "ghost-browser/internal/errors"
)

// NormalizeURL validates a user-supplied URL and applies the default scheme when absent.
func NormalizeURL(input string) (string, error) {
	raw := strings.TrimSpace(input)
	if raw == "" {
		return "", apperrors.ErrEmptyURL
	}

	if !strings.Contains(raw, "://") {
		raw = "https://" + raw
	}

	parsed, err := url.Parse(raw)
	if err != nil {
		return "", fmt.Errorf("%w: %v", apperrors.ErrInvalidURL, err)
	}

	scheme := strings.ToLower(parsed.Scheme)
	if scheme != "http" && scheme != "https" {
		return "", fmt.Errorf("%w: %s", apperrors.ErrUnsupportedScheme, parsed.Scheme)
	}

	if parsed.Host == "" {
		return "", fmt.Errorf("%w: missing host", apperrors.ErrInvalidURL)
	}

	if parsed.Path == "" {
		parsed.Path = "/"
	}

	return parsed.String(), nil
}
