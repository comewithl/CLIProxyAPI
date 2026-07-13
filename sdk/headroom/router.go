package headroom

import (
	"encoding/json"
	"strings"
)

type contentKind int

const (
	contentKindText contentKind = iota
	contentKindJSON
	contentKindCode
	contentKindLog
)

func routeContent(text string) contentKind {
	trimmed := strings.TrimSpace(text)
	if trimmed == "" {
		return contentKindText
	}
	if json.Valid([]byte(trimmed)) {
		return contentKindJSON
	}
	lower := strings.ToLower(trimmed)
	if strings.HasPrefix(lower, "```json") {
		return contentKindJSON
	}
	if strings.HasPrefix(trimmed, "```") || looksLikeCode(trimmed) {
		return contentKindCode
	}
	if looksLikeLog(trimmed) {
		return contentKindLog
	}
	return contentKindText
}

func looksLikeCode(text string) bool {
	markers := []string{"func ", "class ", "interface ", "type ", "package ", "import ", "return ", "throw ", "def "}
	for _, marker := range markers {
		if strings.Contains(text, marker) {
			return true
		}
	}
	return false
}

func looksLikeLog(text string) bool {
	lines := strings.Split(text, "\n")
	matches := 0
	for _, line := range lines {
		upper := strings.ToUpper(strings.TrimSpace(line))
		if strings.HasPrefix(upper, "INFO ") || strings.HasPrefix(upper, "WARN ") || strings.HasPrefix(upper, "ERROR ") || strings.HasPrefix(upper, "FATAL ") || strings.Contains(upper, " ERROR ") {
			matches++
		}
	}
	return matches >= 2
}
