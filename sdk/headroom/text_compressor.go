package headroom

import "strings"

func compressTextContent(text string, opts CompressionOptions) (string, bool) {
	if isAlreadyCompressed(text) {
		return text, false
	}
	lines := strings.Split(text, "\n")
	if len(lines) <= 3 {
		return compressParagraph(text, opts)
	}
	kept := make([]string, 0, len(lines))
	seen := make(map[string]bool, len(lines))
	omitted := 0
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}
		if seen[trimmed] {
			omitted++
			continue
		}
		seen[trimmed] = true
		if shouldKeepTextLine(trimmed) || i == 0 || i == len(lines)-1 {
			kept = append(kept, line)
			continue
		}
		omitted++
	}
	if omitted > 0 {
		kept = appendWithMiddleMarker(kept, headroomOmittedMarker(pluralizeItems(omitted, "line")))
	}
	compressed := strings.Join(kept, "\n")
	if compressed == "" || len(compressed) >= len(text) {
		return text, false
	}
	return compressed, true
}

func compressParagraph(text string, opts CompressionOptions) (string, bool) {
	sentences := splitSentences(text)
	if len(sentences) <= 2 {
		return text, false
	}
	kept := []string{sentences[0]}
	omitted := 0
	for i := 1; i < len(sentences)-1; i++ {
		if shouldKeepTextLine(sentences[i]) {
			kept = append(kept, sentences[i])
			continue
		}
		omitted++
	}
	kept = append(kept, sentences[len(sentences)-1])
	if omitted > 0 {
		kept = appendWithMiddleMarker(kept, headroomOmittedMarker(pluralizeItems(omitted, "sentence")))
	}
	compressed := strings.Join(kept, " ")
	if len(compressed) >= len(text) {
		return text, false
	}
	return compressed, true
}

func shouldKeepTextLine(line string) bool {
	if strings.HasPrefix(line, "#") || strings.HasPrefix(line, "-") || strings.HasPrefix(line, "*") {
		return true
	}
	upper := strings.ToUpper(line)
	keywords := []string{"ERROR", "FATAL", "WARN", "FAIL", "EXCEPTION", "TODO", "ACTION", "FIXME"}
	for _, keyword := range keywords {
		if strings.Contains(upper, keyword) {
			return true
		}
	}
	return false
}

func splitSentences(text string) []string {
	parts := strings.FieldsFunc(text, func(r rune) bool { return r == '.' || r == '!' || r == '?' })
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed != "" {
			out = append(out, trimmed)
		}
	}
	return out
}

func appendWithMiddleMarker(items []string, marker string) []string {
	if len(items) == 0 {
		return []string{marker}
	}
	if len(items) == 1 {
		return append(items, marker)
	}
	out := append([]string{}, items[:1]...)
	out = append(out, marker)
	out = append(out, items[1:]...)
	return out
}
