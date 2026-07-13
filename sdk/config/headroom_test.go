package config

import "testing"

func TestParseConfigBytesHeadroomPublicAPI(t *testing.T) {
	cfg, err := ParseConfigBytes([]byte(`
headroom:
  enabled: true
  strategy: drop-oldest
  rules:
    - models:
        - name: "gpt-*"
          protocol: openai
      max-input-tokens: 100
`))
	if err != nil {
		t.Fatalf("ParseConfigBytes() error = %v", err)
	}

	var headroomCfg HeadroomConfig = cfg.Headroom
	if !headroomCfg.Enabled {
		t.Fatalf("Headroom.Enabled = false, want true")
	}
	if len(headroomCfg.Rules) != 1 {
		t.Fatalf("len(Headroom.Rules) = %d, want 1", len(headroomCfg.Rules))
	}
	var rule HeadroomRule = headroomCfg.Rules[0]
	if rule.MaxInputTokens != 100 {
		t.Fatalf("HeadroomRule.MaxInputTokens = %d, want 100", rule.MaxInputTokens)
	}
}
