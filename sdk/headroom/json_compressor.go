package headroom

import (
	"encoding/json"
	"strings"
)

func compressJSONContent(text string, opts CompressionOptions) (string, bool) {
	trimmed := strings.TrimSpace(text)
	prefix, suffix := "", ""
	if strings.HasPrefix(strings.ToLower(trimmed), "```json") && strings.HasSuffix(trimmed, "```") {
		newline := strings.Index(trimmed, "\n")
		if newline < 0 {
			return text, false
		}
		prefix = trimmed[:newline+1]
		trimmed = strings.TrimSpace(strings.TrimSuffix(strings.TrimPrefix(trimmed, prefix), "```"))
		suffix = "\n```"
	}

	var value any
	if err := json.Unmarshal([]byte(trimmed), &value); err != nil {
		return text, false
	}
	if opts.Mode == CompressionModeAggressive {
		value = compressJSONValue(value)
	}
	out, err := json.Marshal(value)
	if err != nil {
		return text, false
	}
	compressed := prefix + string(out) + suffix
	return compressed, compressed != text
}

func compressJSONValue(value any) any {
	switch typed := value.(type) {
	case map[string]any:
		out := make(map[string]any, len(typed))
		for key, item := range typed {
			compressed := compressJSONValue(item)
			if isEmptyJSONValue(compressed) {
				continue
			}
			out[key] = compressed
		}
		return out
	case []any:
		items := make([]any, 0, len(typed))
		for _, item := range typed {
			compressed := compressJSONValue(item)
			if isEmptyJSONValue(compressed) {
				continue
			}
			items = append(items, compressed)
		}
		if len(items) > 6 {
			omitted := len(items) - 4
			return []any{items[0], items[1], headroomOmittedMarker(pluralizeItems(omitted, "array item")), items[len(items)-2], items[len(items)-1]}
		}
		return items
	default:
		return value
	}
}

func isEmptyJSONValue(value any) bool {
	switch typed := value.(type) {
	case nil:
		return true
	case string:
		return typed == ""
	case []any:
		return len(typed) == 0
	case map[string]any:
		return len(typed) == 0
	default:
		return false
	}
}

func pluralizeItems(count int, noun string) string {
	if count == 1 {
		return "1 " + noun
	}
	return strconvItoa(count) + " " + noun + "s"
}
