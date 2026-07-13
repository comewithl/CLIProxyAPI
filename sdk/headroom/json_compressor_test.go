package headroom

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestCompressJSONConservativeMinifiesOnly(t *testing.T) {
	input := "{\n  \"name\": \"demo\",\n  \"items\": [1, 2, 3],\n  \"empty\": null\n}"
	out, changed := compressJSONContent(input, CompressionOptions{Mode: CompressionModeConservative})
	if !changed {
		t.Fatalf("changed = false, want true")
	}
	if out != `{"empty":null,"items":[1,2,3],"name":"demo"}` {
		t.Fatalf("out = %q", out)
	}
	var decoded map[string]any
	if err := json.Unmarshal([]byte(out), &decoded); err != nil {
		t.Fatalf("compressed JSON invalid: %v", err)
	}
}

func TestCompressJSONFenceWithoutNewlineIsUnchanged(t *testing.T) {
	input := "```json{\"ok\":true}```"
	out, changed := compressJSONContent(input, CompressionOptions{Mode: CompressionModeConservative})
	if changed {
		t.Fatalf("changed = true, want false")
	}
	if out != input {
		t.Fatalf("out = %q, want original", out)
	}
}

func TestCompressJSONAggressiveSamplesLongArrays(t *testing.T) {
	input := `{"items":["a","b","c","d","e","f","g"],"empty":null}`
	out, changed := compressJSONContent(input, CompressionOptions{Mode: CompressionModeAggressive})
	if !changed {
		t.Fatalf("changed = false, want true")
	}
	if strings.Contains(out, `"empty"`) {
		t.Fatalf("empty field not removed: %s", out)
	}
	if !strings.Contains(out, "headroom: omitted") {
		t.Fatalf("omission marker missing: %s", out)
	}
	var decoded map[string]any
	if err := json.Unmarshal([]byte(out), &decoded); err != nil {
		t.Fatalf("compressed JSON invalid: %v", err)
	}
}
