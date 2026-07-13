package executor

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/router-for-me/CLIProxyAPI/v7/internal/config"
	cliproxyauth "github.com/router-for-me/CLIProxyAPI/v7/sdk/cliproxy/auth"
	cliproxyexecutor "github.com/router-for-me/CLIProxyAPI/v7/sdk/cliproxy/executor"
	sdktranslator "github.com/router-for-me/CLIProxyAPI/v7/sdk/translator"
	"github.com/tidwall/gjson"
)

func TestOpenAICompatExecutorAppliesHeadroomBeforeUpstream(t *testing.T) {
	var gotBody []byte
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		gotBody = body
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"id":"chatcmpl_1","object":"chat.completion","choices":[{"index":0,"message":{"role":"assistant","content":"ok"},"finish_reason":"stop"}],"usage":{"prompt_tokens":1,"completion_tokens":1,"total_tokens":2}}`))
	}))
	defer server.Close()

	executor := NewOpenAICompatExecutor("openai-compatibility", &config.Config{
		Headroom: config.HeadroomConfig{
			Enabled: true,
			Rules: []config.HeadroomRule{
				{Models: []config.PayloadModelRule{{Name: "gpt-*", Protocol: "openai"}}, MaxInputTokens: 20, MinRecentMessages: 1},
			},
		},
	})
	auth := &cliproxyauth.Auth{Attributes: map[string]string{"base_url": server.URL + "/v1", "api_key": "test"}}
	payload := []byte(`{"model":"gpt-5","messages":[{"role":"system","content":"rules"},{"role":"user","content":"old message with enough text to remove"},{"role":"user","content":"latest"}]}`)

	_, err := executor.Execute(context.Background(), auth, cliproxyexecutor.Request{Model: "gpt-5", Payload: payload}, cliproxyexecutor.Options{SourceFormat: sdktranslator.FromString("openai")})
	if err != nil {
		t.Fatalf("Execute error: %v", err)
	}
	if gjson.GetBytes(gotBody, `messages.#(content=="old message with enough text to remove")`).Exists() {
		t.Fatalf("old message still sent upstream: %s", string(gotBody))
	}
	if !gjson.GetBytes(gotBody, `messages.#(content=="latest")`).Exists() {
		t.Fatalf("latest message missing upstream: %s", string(gotBody))
	}
}

func TestOpenAICompatExecutorAppliesHeadroomCompressionBeforeUpstream(t *testing.T) {
	var gotBody []byte
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		gotBody = body
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"id":"chatcmpl_1","object":"chat.completion","choices":[{"index":0,"message":{"role":"assistant","content":"ok"},"finish_reason":"stop"}],"usage":{"prompt_tokens":1,"completion_tokens":1,"total_tokens":2}}`))
	}))
	defer server.Close()

	executor := NewOpenAICompatExecutor("openai-compatibility", &config.Config{
		Headroom: config.HeadroomConfig{
			Enabled: true,
			Rules: []config.HeadroomRule{
				{Models: []config.PayloadModelRule{{Name: "gpt-*", Protocol: "openai"}}, Strategy: "compress", MaxInputTokens: 20, MinCompressionTokens: 1, MinRecentMessages: 1},
			},
		},
	})
	auth := &cliproxyauth.Auth{Attributes: map[string]string{"base_url": server.URL + "/v1", "api_key": "test"}}
	payload := []byte(`{"model":"gpt-5","messages":[{"role":"user","content":"# Notes\nnoise line repeated\nnoise line repeated\nWARN important warning\nnoise line repeated\nERROR important failure\nFinal action remains"},{"role":"user","content":"latest"}]}`)

	_, err := executor.Execute(context.Background(), auth, cliproxyexecutor.Request{Model: "gpt-5", Payload: payload}, cliproxyexecutor.Options{SourceFormat: sdktranslator.FromString("openai")})
	if err != nil {
		t.Fatalf("Execute error: %v", err)
	}
	content := gjson.GetBytes(gotBody, "messages.0.content").String()
	if !strings.Contains(content, "headroom: omitted") {
		t.Fatalf("compressed marker missing upstream: %s", string(gotBody))
	}
	if !gjson.GetBytes(gotBody, `messages.#(content=="latest")`).Exists() {
		t.Fatalf("latest message missing upstream: %s", string(gotBody))
	}
}

func TestClaudeExecutorAppliesHeadroomBeforeUpstream(t *testing.T) {
	var gotBody []byte
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		gotBody = body
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"id":"msg_1","type":"message","model":"claude-3-5-sonnet","role":"assistant","content":[{"type":"text","text":"ok"}],"usage":{"input_tokens":1,"output_tokens":1}}`))
	}))
	defer server.Close()

	executor := NewClaudeExecutor(&config.Config{
		Headroom: config.HeadroomConfig{
			Enabled: true,
			Rules: []config.HeadroomRule{
				{Models: []config.PayloadModelRule{{Name: "claude-*", Protocol: "claude"}}, MaxInputTokens: 20, MinRecentMessages: 1},
			},
		},
	})
	auth := &cliproxyauth.Auth{Attributes: map[string]string{"api_key": "key-123", "base_url": server.URL}}
	payload := []byte(`{"model":"claude-3-5-sonnet","messages":[{"role":"user","content":"old message with enough text to remove"},{"role":"user","content":"latest"}]}`)

	_, err := executor.Execute(context.Background(), auth, cliproxyexecutor.Request{Model: "claude-3-5-sonnet", Payload: payload}, cliproxyexecutor.Options{SourceFormat: sdktranslator.FromString("claude")})
	if err != nil {
		t.Fatalf("Execute error: %v", err)
	}
	if gjson.GetBytes(gotBody, `messages.#(content=="old message with enough text to remove")`).Exists() {
		t.Fatalf("old message still sent upstream: %s", string(gotBody))
	}
	if !gjson.GetBytes(gotBody, `messages.#(content=="latest")`).Exists() {
		t.Fatalf("latest message missing upstream: %s", string(gotBody))
	}
}
