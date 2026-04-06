package renderer

import (
	"fmt"
	"strings"

	"ghost-browser/pkg/types"
)

const defaultWidth = 88

type Options struct {
	Width           int
	ReadabilityMode bool
	ShowHelpHint    bool
}

func Render(page *types.Page, opts Options) string {
	if page == nil {
		return "No page loaded.\n"
	}

	width := opts.Width
	if width <= 0 {
		width = defaultWidth
	}

	var lines []string
	lines = append(lines, "Ghost Mode Browser")
	lines = append(lines, strings.Repeat("=", min(width, 19)))
	if page.Title != "" {
		lines = append(lines, "Title: "+page.Title)
	}
	lines = append(lines, "URL: "+page.FinalURL)
	if opts.ReadabilityMode {
		lines = append(lines, "Mode: readability")
	} else {
		lines = append(lines, "Mode: standard")
	}
	lines = append(lines, "")

	content := page.TextContent
	if opts.ReadabilityMode && strings.TrimSpace(page.ReadabilityContent) != "" {
		content = page.ReadabilityContent
	}
	if strings.TrimSpace(content) == "" {
		lines = append(lines, "No readable text extracted.")
	} else {
		for _, paragraph := range strings.Split(content, "\n\n") {
			paragraph = strings.TrimSpace(paragraph)
			if paragraph == "" {
				continue
			}
			lines = append(lines, wrapParagraph(paragraph, width)...)
			lines = append(lines, "")
		}
	}

	if len(page.Warnings) > 0 {
		for _, warning := range page.Warnings {
			lines = append(lines, "Warning: "+warning)
		}
		lines = append(lines, "")
	}

	lines = append(lines, "Links")
	lines = append(lines, strings.Repeat("-", min(width, 5)))
	if len(page.Links) == 0 {
		lines = append(lines, "No links found.")
	} else {
		for _, link := range page.Links {
			lines = append(lines, fmt.Sprintf("[%d] %s", link.Index, link.Label))
			lines = append(lines, "    "+link.URL)
		}
	}

	if opts.ShowHelpHint {
		lines = append(lines, "")
		lines = append(lines, "Commands: open <n>, back, forward, reload, bookmark add, bookmark list, history, search <query>, readability, help, quit")
	}

	return strings.TrimRight(strings.Join(lines, "\n"), "\n") + "\n"
}

func wrapParagraph(text string, width int) []string {
	if width < 20 {
		width = defaultWidth
	}

	words := strings.Fields(text)
	if len(words) == 0 {
		return nil
	}

	var lines []string
	current := words[0]
	for _, word := range words[1:] {
		if len(current)+1+len(word) > width {
			lines = append(lines, current)
			current = word
			continue
		}
		current += " " + word
	}
	lines = append(lines, current)
	return lines
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
