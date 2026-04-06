package parser

import (
	"bytes"
	"fmt"
	"strings"

	"ghost-browser/internal/cleaner"
	apperrors "ghost-browser/internal/errors"
	"ghost-browser/internal/readability"
	"ghost-browser/internal/resolver"
	"ghost-browser/pkg/types"
	"golang.org/x/net/html"
)

type Parser struct{}

func New() *Parser {
	return &Parser{}
}

func (p *Parser) Parse(sourceURL string, fetchResult *types.FetchResult) (*types.Page, error) {
	if fetchResult == nil {
		return nil, fmt.Errorf("%w: missing fetch result", apperrors.ErrInvalidHTML)
	}

	doc, err := html.Parse(bytes.NewReader(fetchResult.Body))
	if err != nil {
		return nil, fmt.Errorf("%w: %v", apperrors.ErrInvalidHTML, err)
	}

	page := &types.Page{
		SourceURL:          sourceURL,
		FinalURL:           fetchResult.FinalURL,
		Title:              extractTitle(fetchResult.Body, doc),
		TextContent:        extractText(doc),
		ReadabilityContent: readability.Extract(doc),
		Links:              resolver.ExtractLinks(doc, fetchResult.FinalURL),
		Metadata: map[string]string{
			"content_type": fetchResult.ContentType,
		},
	}
	page.ArticleLinks, page.UtilityLinks = resolver.SplitLinks(page.Links)
	if page.TextContent == "" {
		page.Warnings = append(page.Warnings, "No readable text extracted from the page.")
	}

	return page, nil
}

func extractTitle(body []byte, root *html.Node) string {
	tokenizer := html.NewTokenizer(bytes.NewReader(body))
	insideTitle := false
	var builder strings.Builder

	for {
		switch tokenizer.Next() {
		case html.ErrorToken:
			if builder.Len() > 0 {
				return cleanTitle(builder.String())
			}
			return extractTitleFromTree(root)
		case html.StartTagToken:
			token := tokenizer.Token()
			if strings.EqualFold(token.Data, "title") {
				insideTitle = true
			} else if insideTitle {
				return cleanTitle(builder.String())
			}
		case html.EndTagToken:
			token := tokenizer.Token()
			if strings.EqualFold(token.Data, "title") {
				return cleanTitle(builder.String())
			}
		case html.TextToken:
			if insideTitle {
				builder.Write(tokenizer.Text())
			}
		}
	}
}

func extractTitleFromTree(root *html.Node) string {
	titleNode := findFirstElement(root, "title")
	if titleNode == nil {
		return ""
	}
	return cleanTitle(textContent(titleNode))
}

func extractText(root *html.Node) string {
	body := findFirstElement(root, "body")
	if body == nil {
		body = root
	}

	var paragraphs []string
	var current strings.Builder

	flushParagraph := func() {
		text := normalizeInlineWhitespace(current.String())
		current.Reset()
		if text == "" {
			return
		}
		if len(paragraphs) > 0 && paragraphs[len(paragraphs)-1] == text {
			return
		}
		paragraphs = append(paragraphs, text)
	}

	var walk func(*html.Node)
	walk = func(node *html.Node) {
		if node == nil {
			return
		}

		switch node.Type {
		case html.TextNode:
			text := normalizeInlineWhitespace(node.Data)
			if text == "" {
				return
			}
			if current.Len() > 0 {
				current.WriteByte(' ')
			}
			current.WriteString(text)
			return
		case html.ElementNode:
			if cleaner.ShouldDropElement(node.Data) || cleaner.IsHiddenElement(node) {
				return
			}
			if node.Data == "br" {
				flushParagraph()
				return
			}
			if cleaner.IsBlockElement(node.Data) {
				flushParagraph()
				for child := node.FirstChild; child != nil; child = child.NextSibling {
					walk(child)
				}
				flushParagraph()
				return
			}
		}

		for child := node.FirstChild; child != nil; child = child.NextSibling {
			walk(child)
		}
	}

	walk(body)
	flushParagraph()

	return strings.Join(paragraphs, "\n\n")
}

func findFirstElement(node *html.Node, tag string) *html.Node {
	if node == nil {
		return nil
	}
	if node.Type == html.ElementNode && strings.EqualFold(node.Data, tag) {
		return node
	}
	for child := node.FirstChild; child != nil; child = child.NextSibling {
		if found := findFirstElement(child, tag); found != nil {
			return found
		}
	}
	return nil
}

func textContent(node *html.Node) string {
	var builder strings.Builder
	var walk func(*html.Node)
	walk = func(current *html.Node) {
		if current == nil {
			return
		}
		if current.Type == html.TextNode {
			builder.WriteString(current.Data)
		}
		for child := current.FirstChild; child != nil; child = child.NextSibling {
			walk(child)
		}
	}
	walk(node)
	return builder.String()
}

func normalizeInlineWhitespace(value string) string {
	return strings.Join(strings.Fields(value), " ")
}

func cleanTitle(value string) string {
	if idx := strings.Index(value, "<"); idx >= 0 {
		value = value[:idx]
	}
	return normalizeInlineWhitespace(value)
}
