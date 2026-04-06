package cleaner

import (
	"strings"

	"golang.org/x/net/html"
)

var droppedElements = map[string]struct{}{
	"iframe":   {},
	"noscript": {},
	"script":   {},
	"style":    {},
	"svg":      {},
}

var blockElements = map[string]struct{}{
	"address":    {},
	"article":    {},
	"aside":      {},
	"blockquote": {},
	"body":       {},
	"dd":         {},
	"div":        {},
	"dl":         {},
	"dt":         {},
	"fieldset":   {},
	"figcaption": {},
	"figure":     {},
	"footer":     {},
	"form":       {},
	"h1":         {},
	"h2":         {},
	"h3":         {},
	"h4":         {},
	"h5":         {},
	"h6":         {},
	"header":     {},
	"hr":         {},
	"li":         {},
	"main":       {},
	"nav":        {},
	"ol":         {},
	"p":          {},
	"pre":        {},
	"section":    {},
	"table":      {},
	"td":         {},
	"th":         {},
	"tr":         {},
	"ul":         {},
}

func ShouldDropElement(tag string) bool {
	_, ok := droppedElements[strings.ToLower(tag)]
	return ok
}

func IsBlockElement(tag string) bool {
	_, ok := blockElements[strings.ToLower(tag)]
	return ok
}

func IsHiddenElement(node *html.Node) bool {
	if node == nil || node.Type != html.ElementNode {
		return false
	}

	for _, attr := range node.Attr {
		switch strings.ToLower(attr.Key) {
		case "hidden":
			return true
		case "aria-hidden":
			if strings.EqualFold(strings.TrimSpace(attr.Val), "true") {
				return true
			}
		case "style":
			style := strings.ToLower(strings.ReplaceAll(attr.Val, " ", ""))
			if strings.Contains(style, "display:none") || strings.Contains(style, "visibility:hidden") {
				return true
			}
		}
	}

	return false
}
