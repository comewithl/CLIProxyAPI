package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseConfigBytes_HeadroomDefaults(t *testing.T) {
	cfg, err := ParseConfigBytes([]byte(`port: 8317`))
	if err != nil {
		t.Fatalf("ParseConfigBytes() error = %v", err)
	}
	if cfg.Headroom.Enabled {
		t.Fatalf("Headroom.Enabled = true, want false")
	}
	if cfg.Headroom.Strategy != "drop-oldest" {
		t.Fatalf("Headroom.Strategy = %q, want drop-oldest", cfg.Headroom.Strategy)
	}
	if cfg.Headroom.MinRecentMessages != 2 {
		t.Fatalf("Headroom.MinRecentMessages = %d, want 2", cfg.Headroom.MinRecentMessages)
	}
}

func TestParseConfigBytes_HeadroomRules(t *testing.T) {
	cfg, err := ParseConfigBytes([]byte(`
headroom:
  enabled: true
  strategy: drop-oldest
  max-input-tokens: -1
  reserve-output-tokens: -2
  min-recent-messages: -3
  rules:
    - models:
        - name: "gpt-*"
          protocol: openai
      max-input-tokens: 120000
      reserve-output-tokens: 4096
      compression-mode: aggressive
      min-compression-tokens: 64
      target-compression-ratio: 0.5
      preserve-code-blocks: true
      preserve-json: true
      stable-output: true
      max-compression-passes: 2
`))
	if err != nil {
		t.Fatalf("ParseConfigBytes() error = %v", err)
	}
	if !cfg.Headroom.Enabled {
		t.Fatalf("Headroom.Enabled = false, want true")
	}
	if cfg.Headroom.MaxInputTokens != 0 || cfg.Headroom.ReserveOutputTokens != 0 || cfg.Headroom.MinRecentMessages != 0 {
		t.Fatalf("negative values were not sanitized: %+v", cfg.Headroom)
	}
	if len(cfg.Headroom.Rules) != 1 {
		t.Fatalf("len(Headroom.Rules) = %d, want 1", len(cfg.Headroom.Rules))
	}
	rule := cfg.Headroom.Rules[0]
	if got := rule.Models[0].Name; got != "gpt-*" {
		t.Fatalf("rule model name = %q, want gpt-*", got)
	}
	if got := rule.Strategy; got != "drop-oldest" {
		t.Fatalf("rule strategy = %q, want drop-oldest", got)
	}
	if got := rule.CompressionMode; got != "aggressive" {
		t.Fatalf("rule compression mode = %q, want aggressive", got)
	}
	if rule.MinCompressionTokens != 64 || rule.TargetCompressionRatio != 0.5 || rule.MaxCompressionPasses != 2 {
		t.Fatalf("rule compression fields not parsed: %+v", rule)
	}
	if !rule.PreserveCodeBlocks || !rule.PreserveJSON || !rule.StableOutput {
		t.Fatalf("rule preserve flags not parsed: %+v", rule)
	}
}

func TestLoadConfigOptional_HeadroomSanitizeMatchesParse(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.yaml")
	content := []byte(`
headroom:
  enabled: true
  max-input-tokens: -1
  min-recent-messages: 0
`)
	if err := os.WriteFile(configPath, content, 0600); err != nil {
		t.Fatalf("write config: %v", err)
	}
	cfg, err := LoadConfigOptional(configPath, false)
	if err != nil {
		t.Fatalf("LoadConfigOptional() error = %v", err)
	}
	if cfg.Headroom.Strategy != "drop-oldest" {
		t.Fatalf("Headroom.Strategy = %q, want drop-oldest", cfg.Headroom.Strategy)
	}
	if cfg.Headroom.MaxInputTokens != 0 {
		t.Fatalf("Headroom.MaxInputTokens = %d, want 0", cfg.Headroom.MaxInputTokens)
	}
	if cfg.Headroom.MinRecentMessages != 2 {
		t.Fatalf("Headroom.MinRecentMessages = %d, want 2", cfg.Headroom.MinRecentMessages)
	}
}
