package cleaner

import (
	"strings"
	"testing"

	"golang.org/x/net/html"
)

func TestShouldSuppressNodeByTag(t *testing.T) {
	t.Parallel()

	node := &html.Node{Type: html.ElementNode, Data: "nav"}
	if !ShouldSuppressNode(node) {
		t.Fatal("expected nav element to be suppressed")
	}
}

func TestShouldSuppressNodeByClassAndID(t *testing.T) {
	t.Parallel()

	tests := []*html.Node{
		{Type: html.ElementNode, Data: "div", Attr: []html.Attribute{{Key: "class", Val: "site-sidebar ad-banner"}}},
		{Type: html.ElementNode, Data: "section", Attr: []html.Attribute{{Key: "id", Val: "newsletter-modal"}}},
		{Type: html.ElementNode, Data: "div", Attr: []html.Attribute{{Key: "role", Val: "navigation toolbar"}}},
	}

	for _, node := range tests {
		if !ShouldSuppressNode(node) {
			t.Fatalf("expected node with attrs %+v to be suppressed", node.Attr)
		}
	}
}

func TestShouldSuppressNodeHidden(t *testing.T) {
	t.Parallel()

	tests := []*html.Node{
		{Type: html.ElementNode, Data: "div", Attr: []html.Attribute{{Key: "hidden", Val: ""}}},
		{Type: html.ElementNode, Data: "div", Attr: []html.Attribute{{Key: "aria-hidden", Val: "true"}}},
		{Type: html.ElementNode, Data: "div", Attr: []html.Attribute{{Key: "style", Val: "display: none;"}}},
	}

	for _, node := range tests {
		if !ShouldSuppressNode(node) {
			t.Fatalf("expected hidden node %+v to be suppressed", node.Attr)
		}
	}
}

func TestMalformedHTMLSafety(t *testing.T) {
	t.Parallel()

	doc, err := html.Parse(strings.NewReader("<html><body><div class='cookie-banner'><p>broken"))
	if err != nil {
		t.Fatalf("html.Parse: %v", err)
	}

	target := firstElementWithAttr(doc, "class", "cookie-banner")
	if target == nil {
		t.Fatal("expected target node")
	}
	if !ShouldSuppressNode(target) {
		t.Fatal("expected malformed DOM node to be safely suppressed")
	}
}

func TestArticleContentContainerIsNotSuppressed(t *testing.T) {
	t.Parallel()

	node := &html.Node{
		Type: html.ElementNode,
		Data: "article",
		Attr: []html.Attribute{{Key: "class", Val: "article-content story-body"}},
	}
	if ShouldSuppressNode(node) {
		t.Fatal("did not expect article content container to be suppressed")
	}
}

func firstElementWithAttr(node *html.Node, key, value string) *html.Node {
	if node == nil {
		return nil
	}
	if node.Type == html.ElementNode {
		for _, attr := range node.Attr {
			if strings.EqualFold(attr.Key, key) && attr.Val == value {
				return node
			}
		}
	}
	for child := node.FirstChild; child != nil; child = child.NextSibling {
		if found := firstElementWithAttr(child, key, value); found != nil {
			return found
		}
	}
	return nil
}
