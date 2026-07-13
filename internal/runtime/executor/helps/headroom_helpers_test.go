package helps

import (
	"context"
	"net/http"
	"strings"
	"testing"

	"github.com/router-for-me/CLIProxyAPI/v7/internal/config"
	"github.com/tidwall/gjson"
)

func TestApplyHeadroomConfigWithRequest_DisabledKeepsPayload(t *testing.T) {
	payload := []byte(`{"messages":[{"role":"user","content":"old"},{"role":"user","content":"latest"}]}`)
	cfg := &config.Config{}

	out := ApplyHeadroomConfigWithRequest(context.Background(), cfg, "gpt-5", "openai", "openai", payload, "", "", nil)
	if string(out) != string(payload) {
		t.Fatalf("payload changed while disabled: %s", string(out))
	}
}

func TestApplyHeadroomConfigWithRequest_UsesMatchingRule(t *testing.T) {
	payload := []byte(`{"messages":[{"role":"system","content":"rules"},{"role":"user","content":"old message with enough text to remove"},{"role":"user","content":"latest"}]}`)
	cfg := &config.Config{
		Headroom: config.HeadroomConfig{
			Enabled:           true,
			Strategy:          "drop-oldest",
			MinRecentMessages: 1,
			Rules: []config.HeadroomRule{
				{
					Models:            []config.PayloadModelRule{{Name: "gpt-*", Protocol: "openai"}},
					MaxInputTokens:    15,
					MinRecentMessages: 1,
				},
			},
		},
	}

	out := ApplyHeadroomConfigWithRequest(context.Background(), cfg, "gpt-5", "openai", "openai", payload, "", "", nil)
	if !gjson.GetBytes(out, `messages.#(content=="latest")`).Exists() {
		t.Fatalf("latest message missing: %s", string(out))
	}
	if gjson.GetBytes(out, `messages.#(content=="old message with enough text to remove")`).Exists() {
		t.Fatalf("old message still exists: %s", string(out))
	}
}

func TestApplyHeadroomConfigWithRequest_SkipsUnmatchedRule(t *testing.T) {
	payload := []byte(`{"messages":[{"role":"user","content":"old message with enough text to remove"},{"role":"user","content":"latest"}]}`)
	cfg := &config.Config{
		Headroom: config.HeadroomConfig{
			Enabled: true,
			Rules: []config.HeadroomRule{
				{Models: []config.PayloadModelRule{{Name: "claude-*", Protocol: "claude"}}, MaxInputTokens: 10},
			},
		},
	}

	out := ApplyHeadroomConfigWithRequest(context.Background(), cfg, "gpt-5", "openai", "openai", payload, "", "", nil)
	if string(out) != string(payload) {
		t.Fatalf("payload changed for unmatched rule: %s", string(out))
	}
}

func TestApplyHeadroomConfigWithRequest_HeaderGate(t *testing.T) {
	payload := []byte(`{"messages":[{"role":"user","content":"old message with enough text to remove"},{"role":"user","content":"latest"}]}`)
	cfg := &config.Config{
		Headroom: config.HeadroomConfig{
			Enabled: true,
			Rules: []config.HeadroomRule{
				{Models: []config.PayloadModelRule{{Name: "gpt-*", Protocol: "openai", Headers: map[string]string{"X-Client-Tier": "tenant-*"}}}, MaxInputTokens: 10, MinRecentMessages: 1},
			},
		},
	}
	headers := http.Header{}
	headers.Set("X-Client-Tier", "tenant-alpha")

	out := ApplyHeadroomConfigWithRequest(context.Background(), cfg, "gpt-5", "openai", "openai", payload, "", "", headers)
	if gjson.GetBytes(out, `messages.#(content=="old message with enough text to remove")`).Exists() {
		t.Fatalf("header-matched old message still exists: %s", string(out))
	}

	headers.Set("X-Client-Tier", "other")
	out = ApplyHeadroomConfigWithRequest(context.Background(), cfg, "gpt-5", "openai", "openai", payload, "", "", headers)
	if string(out) != string(payload) {
		t.Fatalf("payload changed for header mismatch: %s", string(out))
	}
}

func TestApplyHeadroomConfigWithRequest_SkipsImagesEndpoint(t *testing.T) {
	payload := []byte(`{"messages":[{"role":"user","content":"old message with enough text to remove"},{"role":"user","content":"latest"}]}`)
	cfg := &config.Config{Headroom: config.HeadroomConfig{Enabled: true, MaxInputTokens: 10}}

	out := ApplyHeadroomConfigWithRequest(context.Background(), cfg, "gpt-5", "openai", "openai", payload, "", "/v1/images/generations", nil)
	if string(out) != string(payload) {
		t.Fatalf("payload changed for images endpoint: %s", string(out))
	}
}

func TestApplyHeadroomConfigWithRequest_RuleCompressionOverridesGlobal(t *testing.T) {
	payload := []byte(`{"messages":[{"role":"user","content":"{\"items\":[\"a\",\"b\",\"c\",\"d\",\"e\",\"f\",\"g\"],\"empty\":null}"}]}`)
	cfg := &config.Config{
		Headroom: config.HeadroomConfig{
			Enabled:                true,
			Strategy:               "compress",
			CompressionMode:        "conservative",
			MaxInputTokens:         5,
			MinCompressionTokens:   1,
			TargetCompressionRatio: 0.5,
			Rules: []config.HeadroomRule{
				{
					Models:          []config.PayloadModelRule{{Name: "gpt-*", Protocol: "openai"}},
					CompressionMode: "aggressive",
				},
			},
		},
	}

	out := ApplyHeadroomConfigWithRequest(context.Background(), cfg, "gpt-5", "openai", "openai", payload, "", "", nil)
	content := gjson.GetBytes(out, "messages.0.content").String()
	if gjson.Get(content, "empty").Exists() {
		t.Fatalf("rule aggressive compression did not remove empty field: %s", string(out))
	}
	if !strings.Contains(content, "headroom: omitted") {
		t.Fatalf("rule aggressive compression marker missing: %s", string(out))
	}
}
