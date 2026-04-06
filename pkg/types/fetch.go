package types

import "net/http"

// FetchResult captures the normalized output of an HTTP fetch.
type FetchResult struct {
	StatusCode  int
	ContentType string
	Body        []byte
	FinalURL    string
	Headers     http.Header
}
