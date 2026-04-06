package cleaner

import (
	"strings"

	"golang.org/x/net/html"
)

var droppedElements = map[string]struct{}{
	"aside":    {},
	"button":   {},
	"footer":   {},
	"form":     {},
	"header":   {},
	"iframe":   {},
	"input":    {},
	"label":    {},
	"nav":      {},
	"noscript": {},
	"select":   {},
	"script":   {},
	"style":    {},
	"svg":      {},
	"template": {},
	"textarea": {},
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

var junkAttributeTerms = []string{
	"nav",
	"menu",
	"footer",
	"header",
	"sidebar",
	"ad",
	"ads",
	"advert",
	"banner",
	"promo",
	"popup",
	"modal",
	"cookie",
	"consent",
	"share",
	"social",
	"subscribe",
	"newsletter",
	"breadcrumb",
	"related",
	"recommend",
	"comment",
	"toolbar",
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

func HasJunkAttributes(node *html.Node) bool {
	if node == nil || node.Type != html.ElementNode {
		return false
	}

	for _, attr := range node.Attr {
		key := strings.ToLower(strings.TrimSpace(attr.Key))
		if key != "class" && key != "id" && key != "role" && key != "aria-label" {
			continue
		}
		value := normalizeForMatch(attr.Val)
		if value == "" {
			continue
		}
		for _, term := range junkAttributeTerms {
			if containsTerm(value, term) {
				return true
			}
		}
	}

	return false
}

func ShouldSuppressNode(node *html.Node) bool {
	if node == nil {
		return false
	}
	if node.Type != html.ElementNode {
		return false
	}
	return ShouldDropElement(node.Data) || IsHiddenElement(node) || HasJunkAttributes(node)
}

func normalizeForMatch(value string) string {
	value = strings.ToLower(value)
	replacer := strings.NewReplacer("-", " ", "_", " ", "/", " ", ".", " ", ":", " ")
	return strings.Join(strings.Fields(replacer.Replace(value)), " ")
}

func containsTerm(value, term string) bool {
	if value == term {
		return true
	}
	return strings.Contains(" "+value+" ", " "+term+" ")
}
