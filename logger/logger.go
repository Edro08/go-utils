package logger

import (
	"os"
)

// ILogger
// ------------------------------------------------------------------------------------------------
type ILogger interface {
	Info(title string, keys ...any)
	Warn(title string, keys ...any)
	Error(title string, keys ...any)
	Fatal(title string, keys ...any)
	Debug(title string, keys ...any)
}

// ------------------------------------------------------------------------------------------------
// Implementation Methods
// ------------------------------------------------------------------------------------------------

func (l *Logger) Info(title string, keys ...any) {
	l.write(LevelInfo, title, keys...)
}

func (l *Logger) Warn(title string, keys ...any) {
	l.write(LevelWarn, title, keys...)
}

func (l *Logger) Error(title string, keys ...any) {
	l.write(LevelError, title, keys...)
}

func (l *Logger) Fatal(title string, keys ...any) {
	l.write(LevelFatal, title, keys...)
	os.Exit(1)
}

func (l *Logger) Debug(title string, keys ...any) {
	l.write(LevelDebug, title, keys...)
}
