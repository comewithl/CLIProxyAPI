package headroom

import (
	"strings"
	"testing"
)

func TestCompressCodeBlockPreservesFenceAndSignature(t *testing.T) {
	input := "```go\npackage main\n\nimport \"fmt\"\n\nfunc main() {\n\tfmt.Println(\"start\")\n\tfmt.Println(\"middle 1\")\n\tfmt.Println(\"middle 2\")\n\tfmt.Println(\"middle 3\")\n\tfmt.Println(\"end\")\n}\n```"
	out, changed := compressCodeContent(input, CompressionOptions{TargetRatio: 0.5})
	if !changed {
		t.Fatalf("changed = false, want true")
	}
	for _, want := range []string{"```go", "package main", "import \"fmt\"", "func main()", headroomMarkerPrefix} {
		if !strings.Contains(out, want) {
			t.Fatalf("compressed code missing %q: %s", want, out)
		}
	}
}
