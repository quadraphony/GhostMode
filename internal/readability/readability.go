package readability

import (
	"math"
	"strings"

	"ghost-browser/internal/cleaner"
	"golang.org/x/net/html"
)

type Result struct {
	Node  *html.Node
	Score int
	Text  string
}

var (
	positiveContainerTerms = []string{"article", "content", "main", "post", "story", "entry", "body", "docs"}
	negativeContainerTerms = []string{"nav", "menu", "footer", "sidebar", "toolbar", "promo", "related", "comment", "share", "social", "subscribe"}
)

func Extract(root *html.Node) string {
	result := Analyze(root)
	if result.Node == nil {
		return ""
	}
	return result.Text
}

func Analyze(root *html.Node) Result {
	var best *html.Node
	bestScore := 0

	var walk func(*html.Node)
	walk = func(node *html.Node) {
		if node == nil {
			return
		}
		if node.Type == html.ElementNode && !cleaner.ShouldSuppressNode(node) {
			score := scoreNode(node)
			if score > bestScore {
				bestScore = score
				best = node
			}
		}
		for child := node.FirstChild; child != nil; child = child.NextSibling {
			walk(child)
		}
	}

	walk(root)
	if best == nil {
		return Result{}
	}
	return Result{
		Node:  best,
		Score: bestScore,
		Text:  extractText(best),
	}
}

func scoreNode(node *html.Node) int {
	metrics := collectMetrics(node)
	if metrics.textLength < 80 {
		return 0
	}

	score := metrics.textLength
	switch strings.ToLower(node.Data) {
	case "article", "main":
		score += 200
	case "section":
		score += 80
	case "div":
		score += 20
	}

	score += metrics.paragraphCount * 30
	score += metrics.mediumParagraphs * 15
	score += metrics.longParagraphs * 30
	score -= metrics.shortFragments * 18
	score -= int(math.Round(metrics.linkDensity * 200.0))
	score += metrics.textBlockDensity / 6
	score += containerAttributeScore(node)
	score -= metrics.linkCount * 3

	return score
}

type metrics struct {
	textLength       int
	paragraphCount   int
	mediumParagraphs int
	longParagraphs   int
	shortFragments   int
	linkCount        int
	linkDensity      float64
	textBlockDensity int
}

func collectMetrics(node *html.Node) metrics {
	fullText := normalizeWhitespace(textContent(node))
	textLength := len(fullText)
	linkText := normalizeWhitespace(linkOnlyText(node))
	linkTextLength := len(linkText)

	paragraphNodes := countParagraphNodes(node)
	paragraphs := paragraphStats(node)
	linkCount := countElements(node, "a")
	textBlockDensity := 0
	if paragraphNodes > 0 {
		textBlockDensity = textLength / paragraphNodes
	}

	density := 0.0
	if textLength > 0 {
		density = float64(linkTextLength) / float64(textLength)
	}

	return metrics{
		textLength:       textLength,
		paragraphCount:   paragraphNodes,
		mediumParagraphs: paragraphs.medium,
		longParagraphs:   paragraphs.long,
		shortFragments:   paragraphs.short,
		linkCount:        linkCount,
		linkDensity:      density,
		textBlockDensity: textBlockDensity,
	}
}

type paragraphSummary struct {
	medium int
	long   int
	short  int
}

func paragraphStats(node *html.Node) paragraphSummary {
	summary := paragraphSummary{}
	var walk func(*html.Node)
	walk = func(current *html.Node) {
		if current == nil {
			return
		}
		if current.Type == html.ElementNode && (strings.EqualFold(current.Data, "p") || strings.EqualFold(current.Data, "li")) {
			length := len(strings.Fields(normalizeWhitespace(textContent(current))))
			switch {
			case length >= 24:
				summary.long++
			case length >= 12:
				summary.medium++
			case length > 0 && length < 6:
				summary.short++
			}
		}
		for child := current.FirstChild; child != nil; child = child.NextSibling {
			walk(child)
		}
	}
	walk(node)
	return summary
}

func countParagraphNodes(node *html.Node) int {
	return countElements(node, "p")
}

func countElements(node *html.Node, tag string) int {
	count := 0
	var walk func(*html.Node)
	walk = func(current *html.Node) {
		if current == nil {
			return
		}
		if current.Type == html.ElementNode && strings.EqualFold(current.Data, tag) && !cleaner.ShouldSuppressNode(current) {
			count++
		}
		for child := current.FirstChild; child != nil; child = child.NextSibling {
			walk(child)
		}
	}
	walk(node)
	return count
}

func linkOnlyText(node *html.Node) string {
	var b strings.Builder
	var walk func(*html.Node, bool)
	walk = func(current *html.Node, insideLink bool) {
		if current == nil {
			return
		}
		if current.Type == html.ElementNode && strings.EqualFold(current.Data, "a") {
			insideLink = true
		}
		if current.Type == html.TextNode && insideLink {
			b.WriteString(current.Data)
			b.WriteByte(' ')
		}
		for child := current.FirstChild; child != nil; child = child.NextSibling {
			walk(child, insideLink)
		}
	}
	walk(node, false)
	return b.String()
}

func containerAttributeScore(node *html.Node) int {
	score := 0
	for _, attr := range node.Attr {
		if attr.Key != "class" && attr.Key != "id" {
			continue
		}
		value := strings.ToLower(attr.Val)
		for _, term := range positiveContainerTerms {
			if strings.Contains(value, term) {
				score += 60
			}
		}
		for _, term := range negativeContainerTerms {
			if strings.Contains(value, term) {
				score -= 80
			}
		}
	}
	return score
}

func extractText(root *html.Node) string {
	var paragraphs []string
	var current strings.Builder

	flush := func() {
		text := normalizeWhitespace(current.String())
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
			text := normalizeWhitespace(node.Data)
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
			if cleaner.IsBlockElement(node.Data) || strings.EqualFold(node.Data, "br") {
				flush()
			}
		}
		for child := node.FirstChild; child != nil; child = child.NextSibling {
			walk(child)
		}
		if node.Type == html.ElementNode && cleaner.IsBlockElement(node.Data) {
			flush()
		}
	}

	walk(root)
	flush()
	return strings.Join(paragraphs, "\n\n")
}

func textContent(node *html.Node) string {
	var b strings.Builder
	var walk func(*html.Node)
	walk = func(current *html.Node) {
		if current == nil {
			return
		}
		if current.Type == html.TextNode {
			b.WriteString(current.Data)
			b.WriteByte(' ')
		}
		for child := current.FirstChild; child != nil; child = child.NextSibling {
			walk(child)
		}
	}
	walk(node)
	return b.String()
}

func normalizeWhitespace(value string) string {
	return strings.Join(strings.Fields(value), " ")
}
