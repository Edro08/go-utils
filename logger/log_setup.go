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

type Level string

const (
	LevelInfo  Level = "INFO"
	LevelWarn  Level = "WARN"
	LevelError Level = "ERROR"
	LevelFatal Level = "FATAL"
	LevelDebug Level = "DEBUG"
)

var levelOrder = map[Level]int{
	LevelDebug: 0,
	LevelInfo:  1,
	LevelWarn:  2,
	LevelError: 3,
	LevelFatal: 4,
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

func NewLogger(opts Opts) *Logger {
	if opts.Format != FormatJSON && opts.Format != FormatText {
		opts.Format = FormatJSON
	}

	if opts.MinLevel < LevelDebug || opts.MinLevel > LevelFatal {
		opts.MinLevel = LevelDebug
	}

	return &Logger{
		opts: opts,
	}
}
