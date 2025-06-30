package logger

// ------------------------------------------------------------------------------------------------
// Struct Options
// ------------------------------------------------------------------------------------------------

type Opts struct {
	MinLevel Level
	Format   Format
}

// ------------------------------------------------------------------------------------------------
// Level
// ------------------------------------------------------------------------------------------------

type Level int

const (
	DEBUG Level = iota
	INFO
	WARN
	ERROR
	FATAL
)

var levelLabels = map[Level]string{
	DEBUG: "DEBUG",
	INFO:  "INFO",
	WARN:  "WARN",
	ERROR: "ERROR",
	FATAL: "FATAL",
}

// ------------------------------------------------------------------------------------------------
// LogFormat
// ------------------------------------------------------------------------------------------------

type Format string

const (
	FormatJSON Format = "JSON"
	FormatText Format = "TEXT"
)

// ------------------------------------------------------------------------------------------------
// NewLogger
// ------------------------------------------------------------------------------------------------

type Logger struct {
	opts Opts
}

func NewLogger(opts Opts) (*Logger, error) {
	if opts.Format != FormatJSON && opts.Format != FormatText {
		opts.Format = FormatJSON
	}

	if opts.MinLevel < DEBUG || opts.MinLevel > FATAL {
		opts.MinLevel = INFO
	}

	return &Logger{
		opts: opts,
	}, nil
}
