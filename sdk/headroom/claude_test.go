package headroom

import (
	"context"
	"testing"

	"github.com/tidwall/gjson"
)

func TestTrimClaudeUnderBudgetKeepsBody(t *testing.T) {
	body := []byte(`{"system":"rules","messages":[{"role":"user","content":"short"}]}`)

	result, err := TrimClaude(context.Background(), body, Options{MaxInputTokens: 1000})
	if err != nil {
		t.Fatalf("TrimClaude() error = %v", err)
	}
	if result.Changed {
		t.Fatalf("Changed = true, want false")
	}
	if string(result.Body) != string(body) {
		t.Fatalf("Body changed: %s", string(result.Body))
	}
}

func TestTrimClaudeDropsOldestPlainMessages(t *testing.T) {
	body := []byte(`{"system":"rules stay","tools":[{"name":"Bash"}],"messages":[{"role":"user","content":"old message with enough text to remove"},{"role":"assistant","content":"middle message with enough text to remove"},{"role":"user","content":"latest stays"}]}`)

	result, err := TrimClaude(context.Background(), body, Options{MaxInputTokens: 20, MinRecentMessages: 1})
	if err != nil {
		t.Fatalf("TrimClaude() error = %v", err)
	}
	if !result.Changed {
		t.Fatalf("Changed = false, want true")
	}
	if got := gjson.GetBytes(result.Body, "system").String(); got != "rules stay" {
		t.Fatalf("system = %q, want rules stay; body=%s", got, string(result.Body))
	}
	if !gjson.GetBytes(result.Body, "tools.0.name").Exists() {
		t.Fatalf("tools removed: %s", string(result.Body))
	}
	if got := gjson.GetBytes(result.Body, "messages.0.content").String(); got != "latest stays" {
		t.Fatalf("latest message missing, got %q; body=%s", got, string(result.Body))
	}
}

func TestTrimClaudePreservesToolUseAndToolResultMessages(t *testing.T) {
	body := []byte(`{"messages":[{"role":"assistant","content":[{"type":"tool_use","id":"toolu_1","name":"Bash","input":{}}]},{"role":"user","content":[{"type":"tool_result","tool_use_id":"toolu_1","content":"result"}]},{"role":"user","content":"old removable text that should go first"},{"role":"user","content":"latest"}]}`)

	result, err := TrimClaude(context.Background(), body, Options{MaxInputTokens: 18, MinRecentMessages: 1})
	if err != nil {
		t.Fatalf("TrimClaude() error = %v", err)
	}
	if !result.Changed {
		t.Fatalf("Changed = false, want true")
	}
	if !gjson.GetBytes(result.Body, `messages.#(content.#(type=="tool_use"))`).Exists() {
		t.Fatalf("tool_use message removed: %s", string(result.Body))
	}
	if !gjson.GetBytes(result.Body, `messages.#(content.#(type=="tool_result"))`).Exists() {
		t.Fatalf("tool_result message removed: %s", string(result.Body))
	}
	if gjson.GetBytes(result.Body, `messages.#(content=="old removable text that should go first")`).Exists() {
		t.Fatalf("old removable message still exists: %s", string(result.Body))
	}
}

func TestTrimClaudeInvalidJSONReturnsError(t *testing.T) {
	_, err := TrimClaude(context.Background(), []byte(`{"messages":`), Options{MaxInputTokens: 10})
	if err == nil {
		t.Fatalf("TrimClaude() error = nil, want error")
	}
}
