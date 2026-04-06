package readability

import (
	"strings"

	"ghost-browser/internal/cleaner"
	"golang.org/x/net/html"
)

func Extract(root *html.Node) string {
	best := findBestCandidate(root)
	if best == nil {
		return ""
	}
	return extractText(best)
}

func findBestCandidate(root *html.Node) *html.Node {
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
	return best
}

func scoreNode(node *html.Node) int {
	text := normalizeWhitespace(textContent(node))
	if len(text) < 80 {
		return 0
	}

	score := len(text)
	switch strings.ToLower(node.Data) {
	case "article", "main":
		score += 200
	case "section":
		score += 80
	case "div":
		score += 20
	}

	for _, attr := range node.Attr {
		value := strings.ToLower(attr.Val)
		if attr.Key == "class" || attr.Key == "id" {
			if strings.Contains(value, "article") || strings.Contains(value, "content") || strings.Contains(value, "main") || strings.Contains(value, "post") {
				score += 120
			}
			if strings.Contains(value, "nav") || strings.Contains(value, "menu") || strings.Contains(value, "footer") || strings.Contains(value, "sidebar") {
				score -= 120
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
