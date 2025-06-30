package logger

import (
	"encoding/json"
	"fmt"
	"runtime"
	"time"
)

type Entry struct {
	Level     Level     `json:"level"`
	Timestamp time.Time `json:"timestamp"`
	Location  string    `json:"location"`
	Title     string    `json:"title"`
	KeyVals   []any     `json:"keysVals"`
}

func (l *Logger) write(level Level, title string, keysVals ...any) {
	if !(levelOrder[level] >= levelOrder[l.opts.MinLevel]) {
		return
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

func (l *Logger) writeText(e *Entry) {
	line := fmt.Sprintf("%s [%s] (%s) %s%s",
		e.Level,
		e.Timestamp.Format(time.RFC3339),
		e.Location,
		e.Title,
		formatKeys(e.KeyVals),
	)

	fmt.Println(line)
}
