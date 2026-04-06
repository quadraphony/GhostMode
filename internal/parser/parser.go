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

	standardContent := extractPrimaryText(doc)
	readabilityResult := readability.Analyze(doc)

	page := &types.Page{
		SourceURL:          sourceURL,
		FinalURL:           fetchResult.FinalURL,
		Title:              extractTitle(fetchResult.Body, doc),
		TextContent:        standardContent,
		ReadabilityContent: readabilityResult.Text,
		Links:              resolver.ExtractLinks(doc, fetchResult.FinalURL, readabilityResult.Node),
		Metadata: map[string]string{
			"content_type": fetchResult.ContentType,
		},
	}
	page.ArticleLinks, page.UtilityLinks = resolver.SplitLinks(page.Links)
	if page.TextContent == "" {
		page.Warnings = append(page.Warnings, "No readable text extracted from the page.")
	}
	if jsWarning := detectJSShell(doc, page); jsWarning != "" {
		page.Warnings = append(page.Warnings, jsWarning)
		page.Metadata["js_heavy_shell"] = "true"
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
			if cleaner.ShouldSuppressNode(node) {
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

func extractPrimaryText(root *html.Node) string {
	body := findFirstElement(root, "body")
	if body == nil {
		body = root
	}

	result := readability.Analyze(body)
	if result.Node != nil {
		text := normalizeInlineWhitespaceForBlocks(result.Text)
		if looksUsableContent(text, result.Score) {
			return text
		}
	}

	return extractText(body)
}

func looksUsableContent(text string, score int) bool {
	if len(strings.TrimSpace(text)) < 80 {
		return false
	}
	if score < 120 {
		return false
	}
	return true
}

func normalizeInlineWhitespaceForBlocks(value string) string {
	parts := strings.Split(value, "\n\n")
	var cleaned []string
	for _, part := range parts {
		part = normalizeInlineWhitespace(part)
		if part != "" {
			cleaned = append(cleaned, part)
		}
	}
	return strings.Join(cleaned, "\n\n")
}

func detectJSShell(root *html.Node, page *types.Page) string {
	if page == nil {
		return ""
	}

	textLength := len(strings.TrimSpace(page.TextContent))
	readabilityLength := len(strings.TrimSpace(page.ReadabilityContent))
	scriptCount := countElement(root, "script")
	iframeCount := countElement(root, "iframe")
	appShellScore := detectShellMarkers(root)

	if textLength > 180 || readabilityLength > 180 {
		return ""
	}
	if scriptCount+iframeCount < 2 && appShellScore == 0 {
		return ""
	}
	return "This page may depend heavily on JavaScript. GhostMode can only show the limited HTML shell returned by the server."
}

func countElement(node *html.Node, tag string) int {
	count := 0
	var walk func(*html.Node)
	walk = func(current *html.Node) {
		if current == nil {
			return
		}
		if current.Type == html.ElementNode && strings.EqualFold(current.Data, tag) {
			count++
		}
		for child := current.FirstChild; child != nil; child = child.NextSibling {
			walk(child)
		}
	}
	walk(node)
	return count
}

func detectShellMarkers(node *html.Node) int {
	score := 0
	var walk func(*html.Node)
	walk = func(current *html.Node) {
		if current == nil {
			return
		}
		if current.Type == html.ElementNode {
			for _, attr := range current.Attr {
				if attr.Key != "id" && attr.Key != "class" {
					continue
				}
				value := strings.ToLower(attr.Val)
				if strings.Contains(value, "__next") || strings.Contains(value, "root") || strings.Contains(value, "app") || strings.Contains(value, "mount") {
					score++
				}
			}
		}
		for child := current.FirstChild; child != nil; child = child.NextSibling {
			walk(child)
		}
	}
	walk(node)
	return score
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
