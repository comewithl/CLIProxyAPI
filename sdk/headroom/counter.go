package headroom

import "unicode/utf8"

type EstimatedTokenCounter struct{}

func (EstimatedTokenCounter) CountText(_ string, text string) int {
	if text == "" {
		return 0
	}
	count := utf8.RuneCountInString(text)/4 + 1
	if count < 1 {
		return 1
	}
	return count
}

func normalizeOptions(opts Options) Options {
	if opts.Strategy == "" {
		opts.Strategy = StrategyDropOldest
	}
	if opts.CompressionMode == "" {
		opts.CompressionMode = CompressionModeAuto
	}
	if opts.MinRecentMessages < 0 {
		opts.MinRecentMessages = 0
	}
	if opts.ReserveOutputTokens < 0 {
		opts.ReserveOutputTokens = 0
	}
	if opts.MinCompressionTokens < 0 {
		opts.MinCompressionTokens = 0
	}
	if opts.TargetCompressionRatio <= 0 || opts.TargetCompressionRatio >= 1 {
		opts.TargetCompressionRatio = 0.65
	}
	if opts.MaxCompressionPasses <= 0 {
		opts.MaxCompressionPasses = 1
	}
	if opts.TokenCounter == nil {
		opts.TokenCounter = EstimatedTokenCounter{}
	}
	return opts
}
