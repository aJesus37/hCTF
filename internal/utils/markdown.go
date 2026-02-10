package utils

import (
	"bytes"
	"html/template"
	"regexp"
	"strings"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/renderer/html"
)

var md = goldmark.New(
	goldmark.WithExtensions(extension.GFM, extension.Table),
	goldmark.WithRendererOptions(html.WithUnsafe()),
)

// RenderMarkdown converts markdown source to HTML template
func RenderMarkdown(source string) template.HTML {
	var buf bytes.Buffer
	if err := md.Convert([]byte(source), &buf); err != nil {
		return template.HTML("<p>Error rendering markdown</p>")
	}
	return template.HTML(buf.String())
}

// StripMarkdown removes markdown syntax to return plain text
func StripMarkdown(source string) string {
	s := source

	// Remove headers (# ## ### etc)
	s = regexp.MustCompile(`(?m)^#+\s+`).ReplaceAllString(s, "")

	// Remove bold (**text**)
	s = regexp.MustCompile(`\*\*(.+?)\*\*`).ReplaceAllString(s, "$1")

	// Remove italic (*text* and _text_)
	s = regexp.MustCompile(`\*(.+?)\*`).ReplaceAllString(s, "$1")
	s = regexp.MustCompile(`_(.+?)_`).ReplaceAllString(s, "$1")

	// Remove inline code (`code`)
	s = regexp.MustCompile("`(.+?)`").ReplaceAllString(s, "$1")

	// Remove links [text](url)
	s = regexp.MustCompile(`\[(.+?)\]\(.+?\)`).ReplaceAllString(s, "$1")

	// Remove code blocks (```code```)
	s = regexp.MustCompile("(?s)```.*?```").ReplaceAllString(s, "")

	// Remove HTML tags
	s = regexp.MustCompile(`<[^>]+>`).ReplaceAllString(s, "")

	// Clean up extra whitespace
	s = strings.TrimSpace(s)

	return s
}
