package marketing

import (
	"html"
	"regexp"
	"strings"
)

var mustachePattern = regexp.MustCompile(`\{\{\{\s*([\w.]+)\s*\}\}\}|\{\{\s*([\w.]+)\s*\}\}`)

func Render(template string, data map[string]string) string {
	return mustachePattern.ReplaceAllStringFunc(template, func(match string) string {
		if strings.HasPrefix(match, "{{{") {
			key := strings.TrimSpace(match[3 : len(match)-3])
			return data[key]
		}
		key := strings.TrimSpace(match[2 : len(match)-2])
		return html.EscapeString(data[key])
	})
}
