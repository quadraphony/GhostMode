package resolver

import (
	"net/url"
	"sort"
	"strings"

	"ghost-browser/internal/cleaner"
	"ghost-browser/pkg/types"
	"golang.org/x/net/html"
)

func ExtractLinks(root *html.Node, baseURL string, mainContent *html.Node) []types.Link {
	base, err := url.Parse(baseURL)
	if err != nil {
		return nil
	}

	var candidates []types.Link
	seenByURL := make(map[string]types.Link)
	seenByLabelURL := make(map[string]struct{})

	var walk func(*html.Node)
	walk = func(node *html.Node) {
		if node == nil {
			return
		}
		if node.Type == html.ElementNode && strings.EqualFold(node.Data, "a") {
			if link, ok := buildLink(node, base, mainContent); ok {
				labelKey := strings.ToLower(link.Label) + "\n" + link.URL
				if _, exists := seenByLabelURL[labelKey]; exists {
					goto Next
				}
				seenByLabelURL[labelKey] = struct{}{}

				if existing, exists := seenByURL[link.URL]; exists {
					if shouldReplaceLink(existing, link) {
						seenByURL[link.URL] = link
					}
				} else {
					seenByURL[link.URL] = link
				}
			}
		}
	Next:
		for child := node.FirstChild; child != nil; child = child.NextSibling {
			walk(child)
		}
	}

	walk(root)
	for _, link := range seenByURL {
		candidates = append(candidates, link)
	}
	sort.SliceStable(candidates, func(i, j int) bool {
		if candidates[i].Category != candidates[j].Category {
			return candidates[i].Category == types.LinkCategoryArticle
		}
		if candidates[i].Score != candidates[j].Score {
			return candidates[i].Score > candidates[j].Score
		}
		if candidates[i].Label != candidates[j].Label {
			return candidates[i].Label < candidates[j].Label
		}
		return candidates[i].URL < candidates[j].URL
	})
	for i := range candidates {
		candidates[i].Index = i + 1
	}
	return candidates
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

func buildLink(node *html.Node, base *url.URL, mainContent *html.Node) (types.Link, bool) {
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
	if shouldDropLabel(label) {
		return types.Link{}, false
	}

	score, category := classifyLink(node, label, resolved.String(), mainContent)
	return types.Link{
		Index:    0,
		Label:    label,
		URL:      resolved.String(),
		Category: category,
		Score:    score,
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

func classifyLink(node *html.Node, label, target string, mainContent *html.Node) (int, types.LinkCategory) {
	text := strings.ToLower(strings.TrimSpace(label))
	target = strings.ToLower(target)
	score := 0

	if looksLikeUtility(text, target) {
		score -= 70
	}
	if looksLikeArticle(node, text, target) {
		score += 90
	}
	if mainContent != nil && isDescendantOf(node, mainContent) {
		score += 120
	}
	if insideJunkContainer(node) {
		score -= 90
	}
	if wordCount(text) >= 5 {
		score += 30
	}
	if len(text) >= 40 {
		score += 20
	}
	if isLowValueLabel(text) {
		score -= 80
	}

	category := types.LinkCategoryUtility
	if score >= 40 {
		category = types.LinkCategoryArticle
	}
	return score, category
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
		"share", "menu",
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

func shouldDropLabel(label string) bool {
	label = normalizeLinkLabel(label)
	if label == "" {
		return true
	}
	if len([]rune(label)) <= 1 {
		return true
	}
	return false
}

func shouldReplaceLink(existing, candidate types.Link) bool {
	if candidate.Score != existing.Score {
		return candidate.Score > existing.Score
	}
	return len(candidate.Label) > len(existing.Label)
}

func isDescendantOf(node, ancestor *html.Node) bool {
	for current := node; current != nil; current = current.Parent {
		if current == ancestor {
			return true
		}
	}
	return false
}

func insideJunkContainer(node *html.Node) bool {
	for current := node.Parent; current != nil; current = current.Parent {
		if cleaner.ShouldSuppressNode(current) {
			return true
		}
		switch strings.ToLower(current.Data) {
		case "nav", "footer", "header", "aside":
			return true
		}
	}
	return false
}

func isLowValueLabel(text string) bool {
	terms := []string{
		"home", "about", "login", "sign in", "privacy", "terms", "share", "subscribe", "menu",
	}
	for _, term := range terms {
		if text == term {
			return true
		}
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
