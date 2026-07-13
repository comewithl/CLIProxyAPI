package headroom

import (
	"context"
	"strings"
	"testing"

	"github.com/tidwall/gjson"
)

func TestTrimCompressReducesOldMessageWithoutDropping(t *testing.T) {
	old := strings.Repeat("alpha beta gamma delta epsilon. ", 20)
	body := []byte(`{"messages":[{"role":"user","content":"` + old + `"},{"role":"user","content":"latest stays untouched"}]}`)

	result, err := TrimOpenAI(context.Background(), body, Options{Strategy: StrategyCompress, MaxInputTokens: 40, MinRecentMessages: 1, MinCompressionTokens: 10})
	if err != nil {
		t.Fatalf("TrimOpenAI() error = %v", err)
	}
	if !result.Changed {
		t.Fatalf("Changed = false, want true")
	}
	if result.RemovedMessages != 0 {
		t.Fatalf("RemovedMessages = %d, want 0", result.RemovedMessages)
	}
	if result.CompressedMessages == 0 {
		t.Fatalf("CompressedMessages = 0, want > 0")
	}
	messages := gjson.GetBytes(result.Body, "messages")
	if got := len(messages.Array()); got != 2 {
		t.Fatalf("messages length = %d, want 2; body=%s", got, string(result.Body))
	}
	if got := messages.Get("1.content").String(); got != "latest stays untouched" {
		t.Fatalf("latest content = %q", got)
	}
	if !strings.Contains(messages.Get("0.content").String(), headroomMarkerPrefix) {
		t.Fatalf("compressed marker missing: %s", messages.Get("0.content").String())
	}
}

func TestTrimCompressThenDropFallsBackToDropOldest(t *testing.T) {
	body := []byte(`{"messages":[{"role":"user","content":"old short removable text"},{"role":"assistant","content":"middle short removable text"},{"role":"user","content":"latest stays"}]}`)

	result, err := TrimOpenAI(context.Background(), body, Options{Strategy: StrategyCompressThenDrop, MaxInputTokens: 8, MinRecentMessages: 1, MinCompressionTokens: 1000})
	if err != nil {
		t.Fatalf("TrimOpenAI() error = %v", err)
	}
	if !result.Changed {
		t.Fatalf("Changed = false, want true")
	}
	if result.RemovedMessages == 0 {
		t.Fatalf("RemovedMessages = 0, want fallback drop")
	}
	if !gjson.GetBytes(result.Body, `messages.#(content=="latest stays")`).Exists() {
		t.Fatalf("latest message missing: %s", string(result.Body))
	}
}

func TestTrimCompressIsIdempotent(t *testing.T) {
	old := strings.Repeat("repeat this paragraph with useful detail. ", 18)
	body := []byte(`{"messages":[{"role":"user","content":"` + old + `"},{"role":"user","content":"latest"}]}`)
	opts := Options{Strategy: StrategyCompress, MaxInputTokens: 35, MinRecentMessages: 1, MinCompressionTokens: 10}

	first, err := TrimOpenAI(context.Background(), body, opts)
	if err != nil {
		t.Fatalf("first TrimOpenAI() error = %v", err)
	}
	second, err := TrimOpenAI(context.Background(), first.Body, opts)
	if err != nil {
		t.Fatalf("second TrimOpenAI() error = %v", err)
	}
	if string(second.Body) != string(first.Body) {
		t.Fatalf("compression is not idempotent\nfirst=%s\nsecond=%s", string(first.Body), string(second.Body))
	}
}
