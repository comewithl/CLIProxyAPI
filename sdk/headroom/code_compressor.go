package headroom

import "strings"

func compressCodeContent(text string, opts CompressionOptions) (string, bool) {
	if isAlreadyCompressed(text) {
		return text, false
	}
	lines := strings.Split(text, "\n")
	if len(lines) <= 6 {
		return text, false
	}
	keep := make([]string, 0, len(lines))
	seenBlank := false
	omitted := 0
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			if seenBlank {
				omitted++
				continue
			}
			seenBlank = true
			keep = append(keep, line)
			continue
		}
		seenBlank = false
		if shouldKeepCodeLine(trimmed) || i < 3 || i >= len(lines)-2 {
			keep = append(keep, line)
			continue
		}
		omitted++
	}
	if omitted == 0 {
		return text, false
	}
	insert := len(keep) / 2
	out := append([]string{}, keep[:insert]...)
	out = append(out, headroomOmittedMarker(pluralizeItems(omitted, "code line")))
	out = append(out, keep[insert:]...)
	compressed := strings.Join(out, "\n")
	if len(compressed) >= len(text) {
		return text, false
	}
	return compressed, true
}

func shouldKeepCodeLine(line string) bool {
	prefixes := []string{"```", "package ", "import ", "func ", "type ", "class ", "interface ", "def "}
	for _, prefix := range prefixes {
		if strings.HasPrefix(line, prefix) {
			return true
		}
	}
	upper := strings.ToUpper(line)
	keywords := []string{"TODO", "FIXME", "SECURITY", "ERROR", "PANIC", "THROW"}
	for _, keyword := range keywords {
		if strings.Contains(upper, keyword) {
			return true
		}
	}
	return false
}
