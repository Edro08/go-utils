package logger

import (
	"encoding/json"
	"fmt"
	"runtime"
	"strings"
	"time"
)

type Entry struct {
	Level     Level       `json:"level"`
	Timestamp time.Time   `json:"timestamp"`
	Location  string      `json:"location"`
	Title     string      `json:"title"`
	KeyVals   interface{} `json:"keysVals"`
}

func (l *Logger) write(level Level, title string, keysVals ...any) {
	if !shouldLog(level, l.opts.MinLevel) {
		return // No loguear si es menor que el nivel mínimo configurado
	}

	_, file, line, ok := runtime.Caller(2)
	if !ok {
		file = "unknown"
	}

	e := Entry{
		Level:     level,
		Timestamp: time.Now(),
		Location:  fmt.Sprintf("%s:%d", file, line),
		Title:     title,
		KeyVals:   keysVals,
	}

	switch l.opts.Format {
	case FormatJSON:
		l.writeJSON(&e)
	case FormatText:
		l.writeText(&e)
	}
}

func (l *Logger) writeJSON(entry *Entry) {
	data, err := json.Marshal(entry)
	if err != nil {
		return
	}

	fmt.Println(string(data))
}

func (l *Logger) writeText(entry *Entry) {
	var keysSlice []any
	if entry.KeyVals != nil {
		if ks, ok := entry.KeyVals.([]any); ok {
			keysSlice = ks
		}
	}

	line := fmt.Sprintf("%s [%s] (%s) %s%s",
		entry.Level,
		entry.Timestamp.Format(time.RFC3339),
		entry.Location,
		entry.Title,
		formatKeys(keysSlice),
	)

	fmt.Println(line)
}

func formatKeys(keys []any) string {
	if len(keys) == 0 {
		return ""
	}

	var sb strings.Builder
	for i := 0; i < len(keys); i += 2 {
		var key string
		var val any = "<missing>"

		// Determinar key
		if i < len(keys) && keys[i] != nil {
			if k, ok := keys[i].(string); ok && k != "" {
				key = k
			} else {
				key = "<nil>"
			}
		} else {
			key = "<nil>"
		}

		// Determinar val
		if i+1 < len(keys) {
			val = keys[i+1]
			if val == nil {
				val = "<nil>"
			}
		}

		_, _ = fmt.Fprintf(&sb, " | %s: %v", key, val)
	}

	return sb.String()
}

func shouldLog(current, min Level) bool {
	return levelOrder[current] >= levelOrder[min]
}
