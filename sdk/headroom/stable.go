package headroom

import "strings"

const headroomMarkerPrefix = "[headroom:"

func headroomOmittedMarker(detail string) string {
	return headroomMarkerPrefix + " omitted " + detail + "]"
}

func isAlreadyCompressed(text string) bool {
	return strings.Contains(text, headroomMarkerPrefix)
}
