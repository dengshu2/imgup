// Package output renders a public URL into one of the supported emit formats.
package output

import "fmt"

// Format is an output rendering style.
type Format string

const (
	// FormatURL emits the bare URL.
	FormatURL Format = "url"
	// FormatMarkdown emits a Markdown image tag.
	FormatMarkdown Format = "md"
	// FormatHTML emits an HTML <img> tag.
	FormatHTML Format = "html"
)

// Parse validates a format string.
func Parse(s string) (Format, error) {
	switch Format(s) {
	case FormatURL, FormatMarkdown, FormatHTML:
		return Format(s), nil
	default:
		return "", fmt.Errorf("invalid format %q (want: url, md, html)", s)
	}
}

// Render formats url according to f. alt is used as the alt text for md/html.
func Render(f Format, url, alt string) string {
	switch f {
	case FormatMarkdown:
		return fmt.Sprintf("![%s](%s)", alt, url)
	case FormatHTML:
		return fmt.Sprintf("<img src=%q alt=%q>", url, alt)
	default:
		return url
	}
}
