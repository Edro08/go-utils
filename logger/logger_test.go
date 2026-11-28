package logger

import (
	"bytes"
	"context"
	"log/slog"
	"strings"
	"testing"
)

type ctxKey string

const traceKey ctxKey = "trace_id"

func newTestLogger(minLevel Level) (*Logger, *bytes.Buffer) {
	buf := &bytes.Buffer{}
	l, _ := New(Options{
		MinLevel: minLevel,
		Format:   FormatText,
		Writer:   buf,
	})
	return l, buf
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

func TestLogger_Levels(t *testing.T) {
	t.Run("Debug written when level enabled", func(t *testing.T) {
		l, buf := newTestLogger(DEBUG)
		l.Debug("DEBUG_EVENT", "key1", "val1")
		if !strings.Contains(buf.String(), "DEBUG_EVENT") || !strings.Contains(buf.String(), "key1=val1") {
			t.Errorf("Debug() output = %q, expected title and keys", buf.String())
		}
	})

	t.Run("Debug omitted when level disabled", func(t *testing.T) {
		l, buf := newTestLogger(INFO)
		l.Debug("DEBUG_EVENT", "key1", "val1")
		if buf.Len() > 0 {
			t.Errorf("Debug() wrote output when disabled: %q", buf.String())
		}
	})

	t.Run("Info, Warn, Error", func(t *testing.T) {
		l, buf := newTestLogger(INFO)
		l.Info("SERVER_START", "port", 8080)
		l.Warn("CACHE_MISS", "key", "user:123")
		l.Error("DB_CONNECTION_FAILED", "host", "127.0.0.1")
		out := buf.String()
		for _, expected := range []string{"SERVER_START", "port=8080", "CACHE_MISS", "DB_CONNECTION_FAILED"} {
			if !strings.Contains(out, expected) {
				t.Errorf("output = %q, missing %q", out, expected)
			}
		}
	})

	t.Run("FATAL label replaces ERROR+4", func(t *testing.T) {
		buf := &bytes.Buffer{}
		l, _ := New(Options{
			MinLevel: INFO,
			Format:   FormatJSON,
			Writer:   buf,
		})
		l.write(FATAL, "CRITICAL_TEST", "code", 500)
		out := buf.String()
		if !strings.Contains(out, `"level":"FATAL"`) {
			t.Errorf("expected level=FATAL, got %q", out)
		}
		if strings.Contains(out, "ERROR+4") {
			t.Errorf("unexpected ERROR+4 in output: %q", out)
		}
	})
}

func TestLogger_AddSource(t *testing.T) {
	t.Run("AddSource true includes trimmed short file path without host root", func(t *testing.T) {
		buf := &bytes.Buffer{}
		l, err := New(Options{
			MinLevel:  INFO,
			Format:    FormatJSON,
			AddSource: true,
			Writer:    buf,
		})
		if err != nil {
			t.Fatalf("New() error: %v", err)
		}
		l.Info("SOURCE_CHECK")
		out := buf.String()
		if !strings.Contains(out, `"file":"logger/logger_test.go"`) {
			t.Errorf("expected trimmed source file logger/logger_test.go, got %q", out)
		}
	})
}

func TestLogger_ContextMethods(t *testing.T) {
	buf := &bytes.Buffer{}
	l, err := New(Options{
		MinLevel: DEBUG,
		Format:   FormatText,
		Writer:   buf,
		ContextExtractor: func(ctx context.Context) []slog.Attr {
			if v, ok := ctx.Value(traceKey).(string); ok && v != "" {
				return []slog.Attr{slog.String("trace_id", v)}
			}
			return nil
		},
	})
	if err != nil {
		t.Fatalf("New() error: %v", err)
	}

	ctx := context.WithValue(context.Background(), traceKey, "tr-abc-123")
	l.DebugCtx(ctx, "CTX_DEBUG")
	l.InfoCtx(ctx, "CTX_INFO", "step", 1)
	l.WarnCtx(ctx, "CTX_WARN")
	var nilCtx context.Context
	l.ErrorCtx(nilCtx, "CTX_NIL_SAFE")

	out := buf.String()
	for _, expected := range []string{"CTX_DEBUG", "CTX_INFO", "CTX_WARN", "CTX_NIL_SAFE", "trace_id=tr-abc-123", "step=1"} {
		if !strings.Contains(out, expected) {
			t.Errorf("context output = %q, missing %q", out, expected)
		}
	}
}

func TestLogger_FatalAndFatalCtx(t *testing.T) {
	t.Run("Fatal logs FATAL, closes resource and exits with 1", func(t *testing.T) {
		l, buf := newTestLogger(INFO)
		cc := &countingCloser{}
		l.closer = cc
		exitCode := 0
		l.exit = func(code int) { exitCode = code }

		l.Fatal("CRITICAL_FAILURE", "code", 500)

		if exitCode != 1 {
			t.Errorf("exitCode = %d, want 1", exitCode)
		}
		if cc.calls != 1 {
			t.Errorf("closer calls = %d, want 1", cc.calls)
		}
		if !strings.Contains(buf.String(), "level=FATAL") || !strings.Contains(buf.String(), "CRITICAL_FAILURE") {
			t.Errorf("unexpected Fatal output: %q", buf.String())
		}
	})

	t.Run("FatalCtx logs with context, closes resource and exits with 1", func(t *testing.T) {
		buf := &bytes.Buffer{}
		l, _ := New(Options{
			MinLevel: INFO,
			Format:   FormatText,
			Writer:   buf,
			ContextExtractor: func(ctx context.Context) []slog.Attr {
				if v, ok := ctx.Value(traceKey).(string); ok {
					return []slog.Attr{slog.String("trace_id", v)}
				}
				return nil
			},
		})
		cc := &countingCloser{}
		l.closer = cc
		exitCode := 0
		l.exit = func(code int) { exitCode = code }

		ctx := context.WithValue(context.Background(), traceKey, "fatal-trace")
		l.FatalCtx(ctx, "CRITICAL_CTX_FAILURE")

		if exitCode != 1 || cc.calls != 1 {
			t.Errorf("exitCode=%d, closer.calls=%d, want 1 and 1", exitCode, cc.calls)
		}
		if !strings.Contains(buf.String(), "trace_id=fatal-trace") || !strings.Contains(buf.String(), "level=FATAL") {
			t.Errorf("unexpected FatalCtx output: %q", buf.String())
		}
	})
}

func TestLogger_SlogAndDerived(t *testing.T) {
	l, buf := newTestLogger(INFO)
	if got := l.Slog(); got == nil {
		t.Error("Slog() returned nil, expected *slog.Logger")
	}

	var iface ILogger = l
	if same := iface.With(); same != l {
		t.Error("With() without args should return receiver")
	}
	if same := iface.WithGroup(""); same != l {
		t.Error("WithGroup(\"\") should return receiver")
	}

	derived := iface.With("service", "billing").WithGroup("meta")
	if derived == nil || derived.Slog() == nil {
		t.Fatal("With().WithGroup() or Slog() returned nil")
	}

	derived.Info("TRANSACTION", "id", "tx-999")
	out := buf.String()
	if !strings.Contains(out, "service=billing") || !strings.Contains(out, "meta.id=tx-999") {
		t.Errorf("derived logger output = %q, missing contextual/grouped attributes", out)
	}

	buf.Reset()
	groupedWith := iface.WithGroup("meta").With("env", "prod")
	groupedWith.Info("DEPLOY", "ok", true)
	if !strings.Contains(buf.String(), "meta.env=prod") || !strings.Contains(buf.String(), "meta.ok=true") {
		t.Errorf("groupedWith output = %q, expected grouped attributes", buf.String())
	}
}

func TestLogger_TrailingAttributeOrder(t *testing.T) {
	buf := &bytes.Buffer{}
	root, err := New(Options{
		MinLevel: INFO,
		Format:   FormatJSON,
		Writer:   buf,
		ContextExtractor: func(ctx context.Context) []slog.Attr {
			if v, ok := ctx.Value(traceKey).(string); ok && v != "" {
				return []slog.Attr{slog.String("traceID", v)}
			}
			return nil
		},
	})
	if err != nil {
		t.Fatalf("New() error: %v", err)
	}

	appLog := root.With("instanceID", "inst-001")
	ctx := context.WithValue(t.Context(), traceKey, "tr-999")
	appLog.InfoCtx(ctx, "MAIN", "status", "Start server in port 8080")

	out := buf.String()
	statusIdx := strings.Index(out, `"status":"Start server in port 8080"`)
	traceIdx := strings.Index(out, `"traceID":"tr-999"`)
	instIdx := strings.Index(out, `"instanceID":"inst-001"}`)

	if statusIdx == -1 || traceIdx == -1 || instIdx == -1 {
		t.Fatalf("unexpected JSON output: %s", out)
	}
	if statusIdx >= traceIdx || traceIdx >= instIdx {
		t.Fatalf("expected order status < traceID < instanceID (at end), got: %s", out)
	}
}
