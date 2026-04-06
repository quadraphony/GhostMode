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
		Index: index,
		Label: label,
		URL:   resolved.String(),
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
