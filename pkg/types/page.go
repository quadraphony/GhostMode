package types

// Page is the normalized internal representation of a fetched HTML document.
type Page struct {
	SourceURL   string
	FinalURL    string
	Title       string
	TextContent string
	Links       []Link
	Metadata    map[string]string
	Warnings    []string
}

// Link describes a navigable page link. Link extraction is added in a later phase.
type Link struct {
	Index int
	Label string
	URL   string
}
