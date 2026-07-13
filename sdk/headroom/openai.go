package headroom

import "context"

func TrimOpenAI(ctx context.Context, body []byte, opts Options) (Result, error) {
	_ = ctx
	opts.Format = FormatOpenAI
	return trimMessages(body, opts, messageRules{
		isAlwaysProtected: isOpenAIProtectedMessage,
		canRemove:         canRemoveOpenAIMessage,
	})
}

func isOpenAIProtectedMessage(msg any) bool {
	role := messageRole(msg)
	return role == "system" || role == "developer"
}

func canRemoveOpenAIMessage(msg any) bool {
	mapped, ok := messageMap(msg)
	if !ok {
		return false
	}
	role := messageRole(msg)
	if role == "system" || role == "developer" || role == "tool" {
		return false
	}
	if _, ok := mapped["tool_calls"]; ok {
		return false
	}
	if _, ok := mapped["function_call"]; ok {
		return false
	}
	return role == "user" || role == "assistant"
}
