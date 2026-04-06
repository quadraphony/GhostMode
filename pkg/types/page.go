package types

// Page is the normalized internal representation of a fetched HTML document.
type Page struct {
	SourceURL          string
	FinalURL           string
	Title              string
	TextContent        string
	ReadabilityContent string
	Links              []Link
	ArticleLinks       []Link
	UtilityLinks       []Link
	Metadata           map[string]string
	Warnings           []string
}

type Link struct {
	Index    int
	Label    string
	URL      string
	Category LinkCategory
	Snippet  string
	Score    int
}

type LinkCategory string

const (
	LinkCategoryArticle LinkCategory = "article"
	LinkCategoryUtility LinkCategory = "utility"
)
