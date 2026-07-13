package headroom

import (
	"context"
	"encoding/json"
	"fmt"
)

func Trim(ctx context.Context, body []byte, opts Options) (Result, error) {
	opts = normalizeOptions(opts)
	switch opts.Format {
	case FormatOpenAI:
		return TrimOpenAI(ctx, body, opts)
	case FormatClaude:
		return TrimClaude(ctx, body, opts)
	default:
		return Result{Body: body, Warnings: []string{"unsupported format"}}, nil
	}
}

func trimMessages(body []byte, opts Options, rules messageRules) (Result, error) {
	opts = normalizeOptions(opts)
	if opts.MaxInputTokens <= 0 {
		return Result{Body: body, Reason: "disabled"}, nil
	}
	budget := opts.MaxInputTokens - opts.ReserveOutputTokens
	if budget <= 0 {
		return Result{Body: body, Warnings: []string{"token budget is not positive"}}, nil
	}

	var root map[string]any
	if err := json.Unmarshal(body, &root); err != nil {
		return Result{Body: body}, fmt.Errorf("parse request body: %w", err)
	}
	before := estimateValueTokens(root, opts)
	if before <= budget {
		return Result{Body: body, BeforeTokens: before, AfterTokens: before, Reason: "within_budget"}, nil
	}

	rawMessages, ok := root["messages"].([]any)
	if !ok || len(rawMessages) == 0 {
		return Result{Body: body, BeforeTokens: before, AfterTokens: before, Warnings: []string{"messages array not found"}}, nil
	}
	if opts.Strategy == StrategyCompress || opts.Strategy == StrategyCompressThenDrop {
		compressed := compressMessages(root, opts, rules)
		if compressed.Changed {
			result, err := marshalCompressionResult(root, opts, before, compressed)
			if err != nil {
				return Result{Body: body}, fmt.Errorf("marshal compressed request body: %w", err)
			}
			if opts.Strategy == StrategyCompress || result.AfterTokens <= budget {
				return result, nil
			}
			dropOpts := opts
			dropOpts.Strategy = StrategyDropOldest
			dropped, err := trimMessages(result.Body, dropOpts, rules)
			if err != nil {
				return result, nil
			}
			dropped.BeforeTokens = before
			dropped.CompressedMessages = result.CompressedMessages
			dropped.CompressedBytes = result.CompressedBytes
			if dropped.Reason != "" {
				dropped.Reason = "compressed_then_" + dropped.Reason
			}
			return dropped, nil
		}
		if opts.Strategy == StrategyCompress {
			return Result{Body: body, BeforeTokens: before, AfterTokens: before, Warnings: compressed.Warnings}, nil
		}
	}
	protected := protectedMessageIndexes(rawMessages, opts.MinRecentMessages, rules)
	removed := 0
	for originalIndex := range rawMessages {
		currentIndex := originalIndex - removed
		if currentIndex < 0 || currentIndex >= len(rawMessages) {
			continue
		}
		if protected[originalIndex] || !rules.canRemove(rawMessages[currentIndex]) {
			continue
		}
		rawMessages = append(rawMessages[:currentIndex], rawMessages[currentIndex+1:]...)
		root["messages"] = rawMessages
		removed++
		if estimateValueTokens(root, opts) <= budget {
			break
		}
	}
	if removed == 0 {
		return Result{Body: body, BeforeTokens: before, AfterTokens: before, Warnings: []string{"no removable messages found"}}, nil
	}

	out, err := json.Marshal(root)
	if err != nil {
		return Result{Body: body}, fmt.Errorf("marshal trimmed request body: %w", err)
	}
	after := estimateValueTokens(root, opts)
	result := Result{Body: out, Changed: true, BeforeTokens: before, AfterTokens: after, RemovedMessages: removed, Reason: "dropped_oldest_messages"}
	if after > budget {
		result.Warnings = append(result.Warnings, "budget exceeded after preserving required messages")
	}
	return result, nil
}

type messageRules struct {
	isAlwaysProtected func(any) bool
	canRemove         func(any) bool
}

func protectedMessageIndexes(messages []any, minRecent int, rules messageRules) map[int]bool {
	protected := make(map[int]bool, len(messages))
	for i, msg := range messages {
		if rules.isAlwaysProtected != nil && rules.isAlwaysProtected(msg) {
			protected[i] = true
		}
	}
	if minRecent <= 0 {
		return protected
	}
	kept := 0
	for i := len(messages) - 1; i >= 0 && kept < minRecent; i-- {
		if protected[i] {
			continue
		}
		protected[i] = true
		kept++
	}
	return protected
}

func estimateValueTokens(value any, opts Options) int {
	switch typed := value.(type) {
	case string:
		return opts.TokenCounter.CountText(opts.Model, typed)
	case []any:
		total := 2
		for _, item := range typed {
			total += estimateValueTokens(item, opts) + 1
		}
		return total
	case map[string]any:
		total := 2
		for key, item := range typed {
			total += opts.TokenCounter.CountText(opts.Model, key) + estimateValueTokens(item, opts) + 2
		}
		return total
	case float64, bool, nil:
		return 1
	default:
		return 1
	}
}

func messageMap(msg any) (map[string]any, bool) {
	mapped, ok := msg.(map[string]any)
	return mapped, ok
}

func messageRole(msg any) string {
	mapped, ok := messageMap(msg)
	if !ok {
		return ""
	}
	role, _ := mapped["role"].(string)
	return role
}
