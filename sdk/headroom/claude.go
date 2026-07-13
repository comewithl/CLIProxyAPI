package headroom

import "context"

func TrimClaude(ctx context.Context, body []byte, opts Options) (Result, error) {
	_ = ctx
	opts.Format = FormatClaude
	return trimMessages(body, opts, messageRules{
		isAlwaysProtected: func(any) bool { return false },
		canRemove:         canRemoveClaudeMessage,
	})
}

func canRemoveClaudeMessage(msg any) bool {
	mapped, ok := messageMap(msg)
	if !ok {
		return false
	}
	if messageHasClaudeToolBlock(mapped["content"]) {
		return false
	}
	role := messageRole(msg)
	return role == "user" || role == "assistant"
}

func messageHasClaudeToolBlock(content any) bool {
	blocks, ok := content.([]any)
	if !ok {
		return false
	}
	for _, block := range blocks {
		mapped, ok := block.(map[string]any)
		if !ok {
			continue
		}
		typeValue, _ := mapped["type"].(string)
		if typeValue == "tool_use" || typeValue == "tool_result" {
			return true
		}
	}
	return false
}
