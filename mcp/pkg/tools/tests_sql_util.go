package tools

import "strings"

func singleQuoteEscape(s string) string {
	if s == "" {
		return s
	}
	return strings.ReplaceAll(s, "'", "''")
}
