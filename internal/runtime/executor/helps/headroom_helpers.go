package helps

import (
	"context"
	"net/http"
	"strings"

	"github.com/router-for-me/CLIProxyAPI/v7/internal/config"
	"github.com/router-for-me/CLIProxyAPI/v7/sdk/headroom"
	log "github.com/sirupsen/logrus"
)

func ApplyHeadroomConfigWithRequest(ctx context.Context, cfg *config.Config, model string, protocol string, fromProtocol string, payload []byte, requestedModel string, requestPath string, headers http.Header) []byte {
	if cfg == nil || !cfg.Headroom.Enabled || len(payload) == 0 || isImagesEndpointRequestPath(requestPath) {
		return payload
	}
	opts, ok := resolveHeadroomOptions(cfg, model, protocol, fromProtocol, payload, requestedModel, headers)
	if !ok || opts.MaxInputTokens <= 0 {
		return payload
	}
	result, err := headroom.Trim(ctx, payload, opts)
	if err != nil {
		LogWithRequestID(ctx).WithError(err).Warn("headroom trim failed; using original payload")
		return payload
	}
	if len(result.Warnings) > 0 {
		LogWithRequestID(ctx).WithFields(log.Fields{
			"warnings": result.Warnings,
			"protocol": protocol,
			"model":    model,
		}).Warn("headroom trim completed with warnings")
	}
	if result.Changed {
		LogWithRequestID(ctx).WithFields(log.Fields{
			"protocol":         protocol,
			"model":            model,
			"before_tokens":    result.BeforeTokens,
			"after_tokens":     result.AfterTokens,
			"removed_messages": result.RemovedMessages,
		}).Debug("headroom trimmed request messages")
	}
	return result.Body
}

func resolveHeadroomOptions(cfg *config.Config, model string, protocol string, fromProtocol string, payload []byte, requestedModel string, headers http.Header) (headroom.Options, bool) {
	base := headroomOptionsFromConfig(cfg.Headroom)
	base.Format = headroomFormat(protocol)
	base.Model = model
	if base.Format == "" {
		return headroom.Options{}, false
	}
	candidates := payloadModelCandidates(model, requestedModel)
	for i := range cfg.Headroom.Rules {
		rule := cfg.Headroom.Rules[i]
		if !payloadModelRulesMatch(rule.Models, protocol, fromProtocol, headers, payload, "", candidates) {
			continue
		}
		merged := base
		if strings.TrimSpace(rule.Strategy) != "" {
			merged.Strategy = headroom.Strategy(strings.TrimSpace(rule.Strategy))
		}
		if rule.MaxInputTokens > 0 {
			merged.MaxInputTokens = rule.MaxInputTokens
		}
		if rule.ReserveOutputTokens > 0 {
			merged.ReserveOutputTokens = rule.ReserveOutputTokens
		}
		if rule.MinRecentMessages > 0 {
			merged.MinRecentMessages = rule.MinRecentMessages
		}
		if strings.TrimSpace(rule.CompressionMode) != "" {
			merged.CompressionMode = headroom.CompressionMode(strings.TrimSpace(rule.CompressionMode))
		}
		if rule.MinCompressionTokens > 0 {
			merged.MinCompressionTokens = rule.MinCompressionTokens
		}
		if rule.TargetCompressionRatio > 0 {
			merged.TargetCompressionRatio = rule.TargetCompressionRatio
		}
		if rule.PreserveCodeBlocks {
			merged.PreserveCodeBlocks = true
		}
		if rule.PreserveJSON {
			merged.PreserveJSON = true
		}
		if rule.StableOutput {
			merged.StableOutput = true
		}
		if rule.MaxCompressionPasses > 0 {
			merged.MaxCompressionPasses = rule.MaxCompressionPasses
		}
		return merged, true
	}
	return base, len(cfg.Headroom.Rules) == 0
}

func headroomOptionsFromConfig(cfg config.HeadroomConfig) headroom.Options {
	return headroom.Options{
		Strategy:               headroom.Strategy(strings.TrimSpace(cfg.Strategy)),
		MaxInputTokens:         cfg.MaxInputTokens,
		ReserveOutputTokens:    cfg.ReserveOutputTokens,
		MinRecentMessages:      cfg.MinRecentMessages,
		CompressionMode:        headroom.CompressionMode(strings.TrimSpace(cfg.CompressionMode)),
		MinCompressionTokens:   cfg.MinCompressionTokens,
		TargetCompressionRatio: cfg.TargetCompressionRatio,
		PreserveCodeBlocks:     cfg.PreserveCodeBlocks,
		PreserveJSON:           cfg.PreserveJSON,
		StableOutput:           cfg.StableOutput,
		MaxCompressionPasses:   cfg.MaxCompressionPasses,
	}
}

func headroomFormat(protocol string) headroom.Format {
	switch strings.ToLower(strings.TrimSpace(protocol)) {
	case "openai":
		return headroom.FormatOpenAI
	case "claude":
		return headroom.FormatClaude
	default:
		return ""
	}
}
