package browser

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"ghost-browser/internal/fetcher"
	"ghost-browser/pkg/types"
	"golang.org/x/net/html"
)

type DuckDuckGoSearch struct {
	client   *http.Client
	endpoint string
}

func NewDuckDuckGoSearch(endpoint string) *DuckDuckGoSearch {
	if strings.TrimSpace(endpoint) == "" {
		endpoint = "https://html.duckduckgo.com/html/"
	}
	return &DuckDuckGoSearch{
		client:   &http.Client{Timeout: 10 * time.Second},
		endpoint: endpoint,
	}
}

func (d *DuckDuckGoSearch) Search(ctx context.Context, query string) ([]types.SearchResult, error) {
	form := url.Values{}
	form.Set("q", query)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, d.endpoint, strings.NewReader(form.Encode()))
	if err != nil {
		return nil, fmt.Errorf("build search request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("User-Agent", fetcher.DefaultUserAgent)

	resp, err := d.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("search request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read search response: %w", err)
	}
	return parseSearchResults(body), nil
}

func parseSearchResults(body []byte) []types.SearchResult {
	doc, err := html.Parse(bytes.NewReader(body))
	if err != nil {
		return nil
	}

	var results []types.SearchResult
	var walk func(*html.Node)
	walk = func(node *html.Node) {
		if node == nil || len(results) >= 10 {
			return
		}
		if node.Type == html.ElementNode && strings.EqualFold(node.Data, "a") {
			class := attrValue(node, "class")
			if strings.Contains(class, "result__a") || strings.Contains(class, "result-link") {
				href := attrValue(node, "href")
				title := strings.TrimSpace(linkText(node))
				if href != "" && title != "" {
					results = append(results, types.SearchResult{
						Title: title,
						URL:   href,
					})
				}
			}
		}
		for child := node.FirstChild; child != nil; child = child.NextSibling {
			walk(child)
		}
	}
	walk(doc)
	return results
}

func attrValue(node *html.Node, key string) string {
	for _, attr := range node.Attr {
		if strings.EqualFold(attr.Key, key) {
			return attr.Val
		}
	}
	return ""
}

func linkText(node *html.Node) string {
	var parts []string
	var walk func(*html.Node)
	walk = func(current *html.Node) {
		if current == nil {
			return
		}
		if current.Type == html.TextNode {
			text := strings.TrimSpace(current.Data)
			if text != "" {
				parts = append(parts, text)
			}
		}
		for child := current.FirstChild; child != nil; child = child.NextSibling {
			walk(child)
		}
	}
	walk(node)
	return strings.Join(parts, " ")
}
