package resolver

import (
	"net/url"
	"strings"

	"ghost-browser/pkg/types"
	"golang.org/x/net/html"
)

func ExtractLinks(root *html.Node, baseURL string) []types.Link {
	base, err := url.Parse(baseURL)
	if err != nil {
		return nil
	}

	var links []types.Link
	seen := make(map[string]struct{})

	var walk func(*html.Node)
	walk = func(node *html.Node) {
		if node == nil {
			return
		}
		if node.Type == html.ElementNode && strings.EqualFold(node.Data, "a") {
			if link, ok := buildLink(node, base, len(links)+1); ok {
				key := link.URL + "\n" + link.Label
				if _, exists := seen[key]; !exists {
					seen[key] = struct{}{}
					links = append(links, link)
				}
			}
		}
		for child := node.FirstChild; child != nil; child = child.NextSibling {
			walk(child)
		}
	}

	walk(root)
	return links
}

func ResolveReference(baseURL, href string) (string, bool) {
	base, err := url.Parse(baseURL)
	if err != nil {
		return "", false
	}
	ref, err := url.Parse(strings.TrimSpace(href))
	if err != nil {
		return "", false
	}
	if ref.Scheme != "" {
		switch strings.ToLower(ref.Scheme) {
		case "http", "https":
		default:
			return "", false
		}
	}
	resolved := base.ResolveReference(ref)
	if resolved.Scheme != "http" && resolved.Scheme != "https" {
		return "", false
	}
	return resolved.String(), true
}

func buildLink(node *html.Node, base *url.URL, index int) (types.Link, bool) {
	href := ""
	for _, attr := range node.Attr {
		if strings.EqualFold(attr.Key, "href") {
			href = strings.TrimSpace(attr.Val)
			break
		}
	}
	if href == "" {
		return types.Link{}, false
	}
	if strings.HasPrefix(strings.ToLower(href), "javascript:") || strings.HasPrefix(strings.ToLower(href), "mailto:") || strings.HasPrefix(strings.ToLower(href), "tel:") {
		return types.Link{}, false
	}

	ref, err := url.Parse(href)
	if err != nil {
		return types.Link{}, false
	}
	resolved := base.ResolveReference(ref)
	if resolved.Scheme != "http" && resolved.Scheme != "https" {
		return types.Link{}, false
	}

	label := normalizeLinkLabel(textContent(node))
	if label == "" {
		label = resolved.String()
	}

	return types.Link{
		Index:    index,
		Label:    label,
		URL:      resolved.String(),
		Category: classifyLink(node, label, resolved.String()),
	}, true
}

func textContent(node *html.Node) string {
	var parts []string
	var walk func(*html.Node)
	walk = func(current *html.Node) {
		if current == nil {
			return
		}
		if current.Type == html.TextNode {
			trimmed := strings.TrimSpace(current.Data)
			if trimmed != "" {
				parts = append(parts, trimmed)
			}
		}
		for child := current.FirstChild; child != nil; child = child.NextSibling {
			walk(child)
		}
	}
	walk(node)
	return strings.Join(parts, " ")
}

func normalizeLinkLabel(value string) string {
	return strings.Join(strings.Fields(value), " ")
}

func SplitLinks(links []types.Link) (articles []types.Link, utility []types.Link) {
	for _, link := range links {
		switch link.Category {
		case types.LinkCategoryArticle:
			articles = append(articles, link)
		default:
			utility = append(utility, link)
		}
	}
	return articles, utility
}

func classifyLink(node *html.Node, label, target string) types.LinkCategory {
	text := strings.ToLower(strings.TrimSpace(label))
	target = strings.ToLower(target)

	if looksLikeUtility(text, target) {
		return types.LinkCategoryUtility
	}
	if looksLikeArticle(node, text, target) {
		return types.LinkCategoryArticle
	}
	return types.LinkCategoryUtility
}

func looksLikeArticle(node *html.Node, text, target string) bool {
	if strings.Count(target, "/") >= 4 && len(text) >= 24 {
		return true
	}
	if strings.Contains(target, "/article") || strings.Contains(target, "/news/") || strings.Contains(target, "/sport/") || strings.Contains(target, "/business/") || strings.Contains(target, "/opinion/") || strings.Contains(target, "/world/") || strings.Contains(target, "/politics/") || strings.Contains(target, "/life/") {
		return true
	}
	if strings.Contains(text, "|") || strings.Contains(text, ":") {
		return true
	}
	if wordCount(text) >= 5 {
		return true
	}
	if ancestorHasSemantic(node, "article", "main") {
		return true
	}
	return false
}

func looksLikeUtility(text, target string) bool {
	if text == "" {
		return true
	}
	if len(text) <= 3 {
		return true
	}
	utilityTerms := []string{
		"home", "news", "sport", "business", "world", "opinion", "life", "travel",
		"food", "wine", "jobs", "subscribe", "sign in", "login", "register",
		"about", "contact", "privacy", "terms", "faq", "weather", "search",
		"markets", "companies", "books", "recipes", "motoring", "politics",
		"good news", "local", "podcast", "video", "live", "more", "read more",
	}
	for _, term := range utilityTerms {
		if text == term {
			return true
		}
	}
	if strings.Contains(target, "/about") || strings.Contains(target, "/privacy") || strings.Contains(target, "/terms") || strings.Contains(target, "/contact") || strings.Contains(target, "/login") || strings.Contains(target, "/subscribe") {
		return true
	}
	return false
}

func wordCount(value string) int {
	return len(strings.Fields(value))
}

func ancestorHasSemantic(node *html.Node, names ...string) bool {
	for current := node.Parent; current != nil; current = current.Parent {
		if current.Type != html.ElementNode {
			continue
		}
		for _, name := range names {
			if strings.EqualFold(current.Data, name) {
				return true
			}
		}
	}
	return false
}
