package headroom

import (
	"strings"
	"testing"
)

func TestCompressTextKeepsImportantLinesAndDropsRepeats(t *testing.T) {
	input := strings.Join([]string{
		"# Incident Report",
		"noise line repeated",
		"noise line repeated",
		"INFO booting worker",
		"WARN retrying upstream request",
		"noise line repeated",
		"ERROR failed to connect to upstream",
		"final action item remains",
	}, "\n")

	out, changed := compressTextContent(input, CompressionOptions{TargetRatio: 0.55})
	if !changed {
		t.Fatalf("changed = false, want true")
	}
	for _, want := range []string{"# Incident Report", "WARN retrying", "ERROR failed", "final action"} {
		if !strings.Contains(out, want) {
			t.Fatalf("compressed text missing %q: %s", want, out)
		}
	}
	if strings.Count(out, "noise line repeated") > 1 {
		t.Fatalf("duplicate line was not collapsed: %s", out)
	}
}
