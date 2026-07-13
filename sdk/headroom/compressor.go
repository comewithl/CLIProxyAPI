package headroom

import (
	"encoding/json"
)

type CompressionOptions struct {
	Mode        CompressionMode
	TargetRatio float64
}

func compressMessages(root map[string]any, opts Options, rules messageRules) Result {
	rawMessages, ok := root["messages"].([]any)
	if !ok || len(rawMessages) == 0 {
		return Result{Warnings: []string{"messages array not found"}}
	}
	protected := protectedMessageIndexes(rawMessages, opts.MinRecentMessages, rules)
	compressedMessages := 0
	compressedBytes := 0
	for i, msg := range rawMessages {
		if protected[i] || !rules.canRemove(msg) {
			continue
		}
		mapped, ok := messageMap(msg)
		if !ok {
			continue
		}
		content, ok := mapped["content"].(string)
		if !ok || content == "" || isAlreadyCompressed(content) {
			continue
		}
		if opts.MinCompressionTokens > 0 && opts.TokenCounter.CountText(opts.Model, content) < opts.MinCompressionTokens {
			continue
		}
		compressed, changed := compressContent(content, CompressionOptions{Mode: opts.CompressionMode, TargetRatio: opts.TargetCompressionRatio})
		if !changed {
			continue
		}
		mapped["content"] = compressed
		compressedMessages++
		compressedBytes += len(content) - len(compressed)
	}
	if compressedMessages == 0 {
		return Result{Warnings: []string{"no compressible messages found"}}
	}
	root["messages"] = rawMessages
	return Result{Changed: true, CompressedMessages: compressedMessages, CompressedBytes: compressedBytes, Reason: "compressed_messages"}
}

func marshalCompressionResult(root map[string]any, opts Options, before int, result Result) (Result, error) {
	out, err := json.Marshal(root)
	if err != nil {
		return Result{}, err
	}
	result.Body = out
	result.BeforeTokens = before
	result.AfterTokens = estimateValueTokens(root, opts)
	return result, nil
}

func compressContent(text string, opts CompressionOptions) (string, bool) {
	if isAlreadyCompressed(text) {
		return text, false
	}
	switch routeContent(text) {
	case contentKindJSON:
		mode := opts.Mode
		if mode == "" || mode == CompressionModeAuto {
			mode = CompressionModeConservative
		}
		return compressJSONContent(text, CompressionOptions{Mode: mode, TargetRatio: opts.TargetRatio})
	case contentKindCode:
		return compressCodeContent(text, opts)
	case contentKindLog, contentKindText:
		return compressTextContent(text, opts)
	default:
		return compressTextContent(text, opts)
	}
}
