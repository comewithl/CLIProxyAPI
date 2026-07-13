package headroom

import (
	"context"
	"testing"

	"github.com/tidwall/gjson"
)

func TestTrimOpenAIUnderBudgetKeepsBody(t *testing.T) {
	body := []byte(`{"messages":[{"role":"system","content":"rules"},{"role":"user","content":"short"}]}`)

	result, err := TrimOpenAI(context.Background(), body, Options{MaxInputTokens: 1000})
	if err != nil {
		t.Fatalf("TrimOpenAI() error = %v", err)
	}
	if result.Changed {
		t.Fatalf("Changed = true, want false")
	}
	if string(result.Body) != string(body) {
		t.Fatalf("Body changed: %s", string(result.Body))
	}
}

func TestTrimOpenAIDropsOldestPlainMessages(t *testing.T) {
	body := []byte(`{"messages":[{"role":"system","content":"rules stay"},{"role":"user","content":"old message with enough text to remove"},{"role":"assistant","content":"middle message with enough text to remove"},{"role":"user","content":"latest stays"}]}`)

	result, err := TrimOpenAI(context.Background(), body, Options{MaxInputTokens: 20, MinRecentMessages: 1})
	if err != nil {
		t.Fatalf("TrimOpenAI() error = %v", err)
	}
	if !result.Changed {
		t.Fatalf("Changed = false, want true")
	}
	messages := gjson.GetBytes(result.Body, "messages")
	if got := messages.Get("0.role").String(); got != "system" {
		t.Fatalf("messages[0].role = %q, want system; body=%s", got, string(result.Body))
	}
	if got := messages.Get("1.content").String(); got != "latest stays" {
		t.Fatalf("latest message missing, got %q; body=%s", got, string(result.Body))
	}
	if result.RemovedMessages == 0 {
		t.Fatalf("RemovedMessages = 0, want > 0")
	}
}

func TestTrimOpenAIPreservesDeveloperAndToolMessages(t *testing.T) {
	body := []byte(`{"messages":[{"role":"developer","content":"dev rules"},{"role":"assistant","content":"tool call","tool_calls":[{"id":"call_1"}]},{"role":"tool","content":"tool result","tool_call_id":"call_1"},{"role":"user","content":"old removable text that should go first"},{"role":"user","content":"latest"}]}`)

	result, err := TrimOpenAI(context.Background(), body, Options{MaxInputTokens: 18, MinRecentMessages: 1})
	if err != nil {
		t.Fatalf("TrimOpenAI() error = %v", err)
	}
	if !result.Changed {
		t.Fatalf("Changed = false, want true")
	}
	if !gjson.GetBytes(result.Body, `messages.#(role=="developer")`).Exists() {
		t.Fatalf("developer message removed: %s", string(result.Body))
	}
	if !gjson.GetBytes(result.Body, `messages.#(role=="tool")`).Exists() {
		t.Fatalf("tool message removed: %s", string(result.Body))
	}
	if gjson.GetBytes(result.Body, `messages.#(content=="old removable text that should go first")`).Exists() {
		t.Fatalf("old removable message still exists: %s", string(result.Body))
	}
}

func TestTrimOpenAIInvalidJSONReturnsError(t *testing.T) {
	_, err := TrimOpenAI(context.Background(), []byte(`{"messages":`), Options{MaxInputTokens: 10})
	if err == nil {
		t.Fatalf("TrimOpenAI() error = nil, want error")
	}
}
