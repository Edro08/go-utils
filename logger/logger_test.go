package logger

import (
	"bytes"
	"log/slog"
	"os"
	"os/exec"
	"strings"
	"testing"
)

func newTestLogger(minLevel Level) (*Logger, *bytes.Buffer) {
	buf := &bytes.Buffer{}
	h := slog.NewTextHandler(buf, &slog.HandlerOptions{
		Level: minLevel.ToSlog(),
	})
	return &Logger{
		opts: Options{
			MinLevel: minLevel,
			Format:   FormatText,
		},
		slog: slog.New(h),
	}, buf
}

func TestLevel_ToSlog(t *testing.T) {
	tests := []struct {
		name string
		l    Level
		want slog.Level
	}{
		{"DEBUG to slog", DEBUG, slog.LevelDebug},
		{"INFO to slog", INFO, slog.LevelInfo},
		{"WARN to slog", WARN, slog.LevelWarn},
		{"ERROR to slog", ERROR, slog.LevelError},
		{"FATAL to slog", FATAL, slog.LevelError + 4},
		{"Default level to slog", Level(99), slog.LevelInfo},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.l.ToSlog(); got != tt.want {
				t.Errorf("ToSlog() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestLogger_Debug(t *testing.T) {
	t.Run("Debug written when level enabled", func(t *testing.T) {
		l, buf := newTestLogger(DEBUG)
		l.Debug("DEBUG_EVENT", "key1", "val1")
		if !strings.Contains(buf.String(), "DEBUG_EVENT") || !strings.Contains(buf.String(), "key1=val1") {
			t.Errorf("Debug() output = %q, expected to contain title and keys", buf.String())
		}
	})

	t.Run("Debug omitted when level disabled", func(t *testing.T) {
		l, buf := newTestLogger(INFO)
		l.Debug("DEBUG_EVENT", "key1", "val1")
		if buf.Len() > 0 {
			t.Errorf("Debug() wrote output when disabled: %q", buf.String())
		}
	})
}

func TestLogger_Info(t *testing.T) {
	l, buf := newTestLogger(INFO)
	l.Info("SERVER_START", "port", 8080)
	out := buf.String()
	if !strings.Contains(out, "SERVER_START") || !strings.Contains(out, "port=8080") {
		t.Errorf("Info() output = %q, want title and key-values", out)
	}
}

func TestLogger_Warn(t *testing.T) {
	l, buf := newTestLogger(WARN)
	l.Warn("CACHE_MISS", "key", "user:123")
	out := buf.String()
	if !strings.Contains(out, "CACHE_MISS") || !strings.Contains(out, "key=user:123") {
		t.Errorf("Warn() output = %q, want title and key-values", out)
	}
}

func TestLogger_Error(t *testing.T) {
	l, buf := newTestLogger(ERROR)
	l.Error("DB_CONNECTION_FAILED", "host", "127.0.0.1")
	out := buf.String()
	if !strings.Contains(out, "DB_CONNECTION_FAILED") || !strings.Contains(out, "host=127.0.0.1") {
		t.Errorf("Error() output = %q, want title and key-values", out)
	}
}

func TestLogger_Fatal(t *testing.T) {
	if os.Getenv("TEST_FATAL_PROCESS") == "1" {
		l, _ := newTestLogger(FATAL)
		l.Fatal("CRITICAL_FAILURE", "code", 500)
		return
	}

	cmd := exec.Command(os.Args[0], "-test.run=TestLogger_Fatal")
	cmd.Env = append(os.Environ(), "TEST_FATAL_PROCESS=1")
	err := cmd.Run()

	if e, ok := err.(*exec.ExitError); ok && !e.Success() {
		// Process exited with non-zero status (expected from os.Exit(1))
		return
	}
	t.Fatalf("TestLogger_Fatal expected process to exit with status 1, got error: %v", err)
}

func TestLogger_Slog(t *testing.T) {
	l, _ := newTestLogger(INFO)
	if got := l.Slog(); got == nil {
		t.Error("Slog() returned nil, expected *slog.Logger")
	}
}

func TestLogger_With(t *testing.T) {
	l, buf := newTestLogger(INFO)
	derived := l.With("service", "billing", "env", "prod")
	if derived == nil {
		t.Fatal("With() returned nil")
	}

	derived.Info("TRANSACTION", "id", "tx-999")
	out := buf.String()
	if !strings.Contains(out, "service=billing") || !strings.Contains(out, "env=prod") || !strings.Contains(out, "id=tx-999") {
		t.Errorf("With() derived logger output = %q, missing contextual attributes", out)
	}
}

func TestLogger_WithGroup(t *testing.T) {
	l, buf := newTestLogger(INFO)
	derived := l.WithGroup("meta")
	if derived == nil {
		t.Fatal("WithGroup() returned nil")
	}

	derived.Info("USER_LOGIN", "id", 42)
	out := buf.String()
	if !strings.Contains(out, "meta.id=42") && !strings.Contains(out, "id=42") {
		t.Errorf("WithGroup() output = %q, expected grouped attributes", out)
	}
}

func TestLogger_write(t *testing.T) {
	tests := []struct {
		name        string
		minLevel    Level
		writeLevel  Level
		title       string
		keysVals    []any
		shouldWrite bool
	}{
		{
			name:        "Write enabled level",
			minLevel:    INFO,
			writeLevel:  INFO,
			title:       "EVENT_OK",
			keysVals:    []any{"status", "ok"},
			shouldWrite: true,
		},
		{
			name:        "Skip disabled level",
			minLevel:    WARN,
			writeLevel:  INFO,
			title:       "EVENT_SKIPPED",
			keysVals:    []any{"status", "skipped"},
			shouldWrite: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l, buf := newTestLogger(tt.minLevel)
			l.write(tt.writeLevel, tt.title, tt.keysVals...)
			if tt.shouldWrite && buf.Len() == 0 {
				t.Errorf("write() did not write output when level was enabled")
			}
			if !tt.shouldWrite && buf.Len() > 0 {
				t.Errorf("write() wrote output when level was disabled: %q", buf.String())
			}
		})
	}
}
