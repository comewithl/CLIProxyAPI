package headroom

import (
	"context"
	"strings"
	"testing"

	"github.com/tidwall/gjson"
)

func TestTrimClaudeCompressPreservesToolMessages(t *testing.T) {
	old := strings.Repeat("old details that can be compressed. ", 16)
	body := []byte(`{"messages":[{"role":"assistant","content":[{"type":"tool_use","id":"toolu_1","name":"Bash","input":{}}]},{"role":"user","content":[{"type":"tool_result","tool_use_id":"toolu_1","content":"result"}]},{"role":"user","content":"` + old + `"},{"role":"user","content":"latest"}]}`)

	result, err := TrimClaude(context.Background(), body, Options{Strategy: StrategyCompress, MaxInputTokens: 40, MinRecentMessages: 1, MinCompressionTokens: 10})
	if err != nil {
		t.Fatalf("TrimClaude() error = %v", err)
	}
	if !gjson.GetBytes(result.Body, `messages.#(content.#(type=="tool_use"))`).Exists() {
		t.Fatalf("tool_use message removed: %s", string(result.Body))
	}
	if !gjson.GetBytes(result.Body, `messages.#(content.#(type=="tool_result"))`).Exists() {
		t.Fatalf("tool_result message removed: %s", string(result.Body))
	}
	if result.CompressedMessages == 0 {
		t.Fatalf("CompressedMessages = 0, want > 0")
	}
}
