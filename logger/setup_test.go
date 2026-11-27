package logger

import (
	"bytes"
	"errors"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
)

type countingCloser struct {
	mu    sync.Mutex
	calls int
	err   error
}

func (c *countingCloser) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.calls++
	return c.err
}

func TestLogger_Close(t *testing.T) {
	t.Run("Nil closer returns nil error", func(t *testing.T) {
		l := &Logger{closer: nil}
		if err := l.Close(); err != nil {
			t.Errorf("Close() error = %v, want nil", err)
		}
	})

	t.Run("Direct closer returns expected error", func(t *testing.T) {
		l := &Logger{closer: &countingCloser{err: errors.New("close failure")}}
		if err := l.Close(); err == nil {
			t.Error("Close() expected error, got nil")
		}
	})

	t.Run("Idempotent onceCloser across derived loggers and concurrent calls", func(t *testing.T) {
		cc := &countingCloser{}
		oc := &onceCloser{closer: cc}
		root := &Logger{
			opts:   Options{MinLevel: INFO, Format: FormatJSON},
			slog:   slog.New(slog.NewJSONHandler(io.Discard, nil)),
			closer: oc,
		}
		derived := root.With("req_id", "123").WithGroup("http")

		var wg sync.WaitGroup
		for i := 0; i < 10; i++ {
			wg.Add(2)
			go func() {
				defer wg.Done()
				_ = root.Close()
			}()
			go func() {
				defer wg.Done()
				_ = derived.Close()
			}()
		}
		wg.Wait()

		if cc.calls != 1 {
			t.Errorf("underlying Close() called %d times, want 1", cc.calls)
		}
	})
}

func TestNew(t *testing.T) {
	tempDir := t.TempDir()
	logFile := filepath.Join(tempDir, "test.log")
	fileOnlyPath := filepath.Join(tempDir, "only-file.log")
	bothPath := filepath.Join(tempDir, "both.log")

	t.Run("Default options uses INFO, JSON, Stdout, AddSource false", func(t *testing.T) {
		buf := &bytes.Buffer{}
		got, err := New(Options{Writer: buf})
		if err != nil {
			t.Fatalf("New() unexpected error: %v", err)
		}
		if got.opts.MinLevel != INFO {
			t.Errorf("MinLevel = %v, want INFO", got.opts.MinLevel)
		}
		if got.opts.Format != FormatJSON {
			t.Errorf("Format = %v, want FormatJSON", got.opts.Format)
		}
		if got.opts.Output != Stdout {
			t.Errorf("Output = %v, want Stdout", got.opts.Output)
		}
		if got.opts.AddSource {
			t.Error("AddSource = true, want false by default")
		}

		// DEBUG must be suppressed by default
		got.Debug("SHOULD_NOT_APPEAR")
		if buf.Len() > 0 {
			t.Errorf("expected DEBUG to be suppressed with default Options{}, got %q", buf.String())
		}

		// INFO must be written without source field
		got.Info("SERVER_READY", "port", 8080)
		out := buf.String()
		if !strings.Contains(out, `"msg":"SERVER_READY"`) {
			t.Errorf("expected msg in JSON output, got %q", out)
		}
		if strings.Contains(out, `"source"`) {
			t.Errorf("expected no source field when AddSource is false, got %q", out)
		}
	})

	t.Run("Backward-compatible EnableFile writes to console and file", func(t *testing.T) {
		buf := &bytes.Buffer{}
		got, err := New(Options{
			EnableFile: true,
			FilePath:   logFile,
			Format:     FormatJSON,
			MinLevel:   DEBUG,
			Writer:     buf,
		})
		if err != nil {
			t.Fatalf("New() unexpected error: %v", err)
		}
		got.Debug("DEBUG_BOTH", "ok", true)
		if err := got.Close(); err != nil {
			t.Fatalf("Close() error: %v", err)
		}
		// Calling Close a second time must not fail
		if err := got.Close(); err != nil {
			t.Fatalf("second Close() error: %v", err)
		}

		data, err := os.ReadFile(logFile)
		if err != nil {
			t.Fatalf("ReadFile() error: %v", err)
		}
		if !strings.Contains(buf.String(), "DEBUG_BOTH") || !strings.Contains(string(data), "DEBUG_BOTH") {
			t.Errorf("expected log in both console (%q) and file (%q)", buf.String(), string(data))
		}
	})

	t.Run("Output File writes only to file and not console", func(t *testing.T) {
		buf := &bytes.Buffer{}
		got, err := New(Options{
			Output:   File,
			FilePath: fileOnlyPath,
			Format:   FormatText,
			MinLevel: INFO,
			Writer:   buf,
		})
		if err != nil {
			t.Fatalf("New() unexpected error: %v", err)
		}
		got.Info("FILE_ONLY_EVENT")
		_ = got.Close()

		if buf.Len() > 0 {
			t.Errorf("expected nothing on console when Output=File, got %q", buf.String())
		}
		data, err := os.ReadFile(fileOnlyPath)
		if err != nil {
			t.Fatalf("ReadFile() error: %v", err)
		}
		if !strings.Contains(string(data), "FILE_ONLY_EVENT") {
			t.Errorf("expected FILE_ONLY_EVENT in file, got %q", string(data))
		}
	})

	t.Run("Output Both writes to console and file", func(t *testing.T) {
		buf := &bytes.Buffer{}
		got, err := New(Options{
			Output:   Both,
			FilePath: bothPath,
			Format:   FormatText,
			MinLevel: INFO,
			Writer:   buf,
		})
		if err != nil {
			t.Fatalf("New() unexpected error: %v", err)
		}
		got.Info("BOTH_DESTINATIONS")
		_ = got.Close()

		data, err := os.ReadFile(bothPath)
		if err != nil {
			t.Fatalf("ReadFile() error: %v", err)
		}
		if !strings.Contains(buf.String(), "BOTH_DESTINATIONS") || !strings.Contains(string(data), "BOTH_DESTINATIONS") {
			t.Errorf("expected output in both console and file")
		}
	})

	t.Run("Enable file logging with directory generates timestamped file", func(t *testing.T) {
		subDir := filepath.Join(tempDir, "tslogs")
		got, err := New(Options{
			Output:   File,
			FilePath: subDir,
			Format:   FormatText,
			MinLevel: INFO,
		})
		if err != nil {
			t.Fatalf("New() error: %v", err)
		}
		got.Info("TS_EVENT")
		_ = got.Close()

		entries, err := os.ReadDir(subDir)
		if err != nil || len(entries) != 1 {
			t.Fatalf("expected 1 log file in %s, got %d (err=%v)", subDir, len(entries), err)
		}
	})
}

func TestParseLevelAndFormat(t *testing.T) {
	levelCases := []struct {
		in     string
		want   Level
		wantOK bool
	}{
		{"debug", DEBUG, true},
		{"INFO", INFO, true},
		{" warn ", WARN, true},
		{"warning", WARN, true},
		{"ERROR", ERROR, true},
		{"Fatal", FATAL, true},
		{"unknown", INFO, false},
		{"", INFO, false},
	}
	for _, tc := range levelCases {
		got, ok := ParseLevel(tc.in)
		if got != tc.want || ok != tc.wantOK {
			t.Errorf("ParseLevel(%q) = (%v, %v), want (%v, %v)", tc.in, got, ok, tc.want, tc.wantOK)
		}
	}

	if DEBUG.String() != "DEBUG" || INFO.String() != "INFO" || WARN.String() != "WARN" || ERROR.String() != "ERROR" || FATAL.String() != "FATAL" || Level(99).String() != "INFO" {
		t.Error("Level.String() returned unexpected label")
	}

	if f, ok := ParseFormat("text"); !ok || f != FormatText {
		t.Errorf("ParseFormat(text) = (%v, %v), want (TEXT, true)", f, ok)
	}
	if f, ok := ParseFormat(" JSON "); !ok || f != FormatJSON {
		t.Errorf("ParseFormat(JSON) = (%v, %v), want (JSON, true)", f, ok)
	}
	if f, ok := ParseFormat("xml"); ok || f != FormatJSON {
		t.Errorf("ParseFormat(xml) = (%v, %v), want (JSON, false)", f, ok)
	}
}

func Test_shortSourceFile(t *testing.T) {
	tests := []struct {
		in   string
		want string
	}{
		{"", ""},
		{"main.go", "main.go"},
		{"cmd/main.go", "cmd/main.go"},
		{"a/b/c/logger.go", "c/logger.go"},
		{`a\b\c\logger.go`, "c/logger.go"},
		{"/main.go", "main.go"},
	}
	for _, tt := range tests {
		if got := shortSourceFile(tt.in); got != tt.want {
			t.Errorf("shortSourceFile(%q) = %q, want %q", tt.in, got, tt.want)
		}
	}
}

func Test_validateOptions(t *testing.T) {
	tests := []struct {
		name       string
		input      Options
		wantFormat Format
		wantMinLvl Level
		wantOutput Output
	}{
		{
			name:       "Empty options defaults to INFO, JSON, Stdout",
			input:      Options{},
			wantFormat: FormatJSON,
			wantMinLvl: INFO,
			wantOutput: Stdout,
		},
		{
			name: "Preserve valid text format and explicit DEBUG level",
			input: Options{
				Format:   FormatText,
				MinLevel: DEBUG,
			},
			wantFormat: FormatText,
			wantMinLvl: DEBUG,
			wantOutput: Stdout,
		},
		{
			name: "EnableFile true maps empty Output to Both",
			input: Options{
				EnableFile: true,
			},
			wantFormat: FormatJSON,
			wantMinLvl: INFO,
			wantOutput: Both,
		},
		{
			name: "Invalid format and out-of-range level reset to defaults",
			input: Options{
				Format:   "XML",
				MinLevel: Level(99),
				Output:   "INVALID",
			},
			wantFormat: FormatJSON,
			wantMinLvl: INFO,
			wantOutput: Stdout,
		},
		{
			name: "Negative level gets reset to INFO",
			input: Options{
				Format:   FormatJSON,
				MinLevel: Level(-5),
			},
			wantFormat: FormatJSON,
			wantMinLvl: INFO,
			wantOutput: Stdout,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts := tt.input
			validateOptions(&opts)
			if opts.Format != tt.wantFormat {
				t.Errorf("validateOptions() Format = %v, want %v", opts.Format, tt.wantFormat)
			}
			if opts.MinLevel != tt.wantMinLvl {
				t.Errorf("validateOptions() MinLevel = %v, want %v", opts.MinLevel, tt.wantMinLvl)
			}
			if opts.Output != tt.wantOutput {
				t.Errorf("validateOptions() Output = %v, want %v", opts.Output, tt.wantOutput)
			}
		})
	}
}

func Test_setupWriter_DefaultDirCleanup(t *testing.T) {
	gotW, gotCloser, err := setupWriter(Options{Output: File, FilePath: ""})
	if err != nil {
		t.Fatalf("setupWriter() error = %v", err)
	}
	if gotW == nil || gotCloser == nil {
		t.Fatal("setupWriter() returned nil writer or closer")
	}
	_ = gotCloser.Close()
	_ = os.RemoveAll("./log")
}
