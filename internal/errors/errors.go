package errors

import stderrors "errors"

var (
	ErrInvalidURL         = stderrors.New("invalid URL")
	ErrUnsupportedScheme  = stderrors.New("unsupported URL scheme")
	ErrUnsupportedContent = stderrors.New("unsupported content type")
	ErrEmptyURL           = stderrors.New("URL is required")
	ErrTooManyRedirects   = stderrors.New("too many redirects")
	ErrEmptyResponseBody  = stderrors.New("empty response body")
	ErrInvalidHTML        = stderrors.New("invalid HTML")
)
