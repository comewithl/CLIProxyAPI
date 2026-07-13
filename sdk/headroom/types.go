package headroom

import "context"

type Format string

const (
	FormatOpenAI Format = "openai"
	FormatClaude Format = "claude"
)

type Strategy string

const (
	StrategyDropOldest       Strategy = "drop-oldest"
	StrategyCompress         Strategy = "compress"
	StrategyCompressThenDrop Strategy = "compress-then-drop"
)

type CompressionMode string

const (
	CompressionModeAuto         CompressionMode = "auto"
	CompressionModeConservative CompressionMode = "conservative"
	CompressionModeAggressive   CompressionMode = "aggressive"
)

type Options struct {
	Format                 Format
	Model                  string
	MaxInputTokens         int
	ReserveOutputTokens    int
	MinRecentMessages      int
	Strategy               Strategy
	CompressionMode        CompressionMode
	MinCompressionTokens   int
	TargetCompressionRatio float64
	PreserveCodeBlocks     bool
	PreserveJSON           bool
	StableOutput           bool
	MaxCompressionPasses   int
	TokenCounter           TokenCounter
}

type TokenCounter interface {
	CountText(model string, text string) int
}

type Result struct {
	Body                []byte
	Changed             bool
	BeforeTokens        int
	AfterTokens         int
	RemovedMessages     int
	CompressedMessages  int
	CompressedBytes     int
	Reason              string
	Warnings            []string
	CompressionWarnings []string
}

type Trimmer struct {
	opts Options
}

func New(opts Options) *Trimmer {
	return &Trimmer{opts: normalizeOptions(opts)}
}

func (t *Trimmer) Trim(ctx context.Context, body []byte) (Result, error) {
	return Trim(ctx, body, t.opts)
}
