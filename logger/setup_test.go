package logger

import (
	"bytes"
	"errors"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"testing"
)

type mockCloser struct {
	closed bool
	err    error
}

func (m *mockCloser) Close() error {
	m.closed = true
	return m.err
}

func TestLogger_Close(t *testing.T) {
	type fields struct {
		closer io.Closer
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name: "Nil closer returns nil error",
			fields: fields{
				closer: nil,
			},
			wantErr: false,
		},
		{
			name: "Valid closer successfully closes",
			fields: fields{
				closer: &mockCloser{err: nil},
			},
			wantErr: false,
		},
		{
			name: "Closer returns error",
			fields: fields{
				closer: &mockCloser{err: errors.New("close failure")},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := &Logger{
				closer: tt.fields.closer,
			}
			if err := l.Close(); (err != nil) != tt.wantErr {
				t.Errorf("Close() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestNew(t *testing.T) {
	tempDir := t.TempDir()
	logFile := filepath.Join(tempDir, "test.log")

	type args struct {
		opts Options
	}
	tests := []struct {
		name       string
		args       args
		wantCloser bool
		wantErr    bool
	}{
		{
			name: "Default options with Stdout",
			args: args{
				opts: Options{},
			},
			wantCloser: false,
			wantErr:    false,
		},
		{
			name: "Enable file logging with specific file",
			args: args{
				opts: Options{
					EnableFile: true,
					FilePath:   logFile,
					Format:     FormatJSON,
					MinLevel:   DEBUG,
				},
			},
			wantCloser: true,
			wantErr:    false,
		},
		{
			name: "Enable file logging with directory only",
			args: args{
				opts: Options{
					EnableFile: true,
					FilePath:   tempDir,
					Format:     FormatText,
					MinLevel:   INFO,
				},
			},
			wantCloser: true,
			wantErr:    false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := New(tt.args.opts)
			if (err != nil) != tt.wantErr {
				t.Errorf("New() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got == nil {
				t.Fatal("New() returned nil logger")
			}
			if got.slog == nil {
				t.Error("New() slog is nil")
			}
			if (got.closer != nil) != tt.wantCloser {
				t.Errorf("New() closer present = %v, want %v", got.closer != nil, tt.wantCloser)
			}
			if got.closer != nil {
				_ = got.Close()
			}
		})
	}
}

func Test_buildHandler(t *testing.T) {
	tests := []struct {
		name       string
		opts       Options
		isJSONType bool
	}{
		{
			name: "JSON format handler",
			opts: Options{
				Format:   FormatJSON,
				MinLevel: INFO,
			},
			isJSONType: true,
		},
		{
			name: "Text format handler",
			opts: Options{
				Format:   FormatText,
				MinLevel: DEBUG,
			},
			isJSONType: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := &bytes.Buffer{}
			got := buildHandler(w, tt.opts)
			if got == nil {
				t.Fatal("buildHandler() returned nil handler")
			}

			if tt.isJSONType {
				if _, ok := got.(*slog.JSONHandler); !ok {
					t.Errorf("buildHandler() got = %T, want *slog.JSONHandler", got)
				}
			} else {
				if _, ok := got.(*slog.TextHandler); !ok {
					t.Errorf("buildHandler() got = %T, want *slog.TextHandler", got)
				}
			}
		})
	}
}

func Test_setupWriter(t *testing.T) {
	tempDir := t.TempDir()
	logFilePath := filepath.Join(tempDir, "app.log")

	type args struct {
		opts Options
	}
	tests := []struct {
		name       string
		args       args
		wantWriter bool
		wantCloser bool
		wantErr    bool
	}{
		{
			name: "EnableFile false returns Stdout without closer",
			args: args{
				opts: Options{EnableFile: false},
			},
			wantWriter: true,
			wantCloser: false,
			wantErr:    false,
		},
		{
			name: "EnableFile true with valid file path",
			args: args{
				opts: Options{
					EnableFile: true,
					FilePath:   logFilePath,
				},
			},
			wantWriter: true,
			wantCloser: true,
			wantErr:    false,
		},
		{
			name: "EnableFile true with directory path",
			args: args{
				opts: Options{
					EnableFile: true,
					FilePath:   tempDir,
				},
			},
			wantWriter: true,
			wantCloser: true,
			wantErr:    false,
		},
		{
			name: "EnableFile true with empty dir defaults to ./log",
			args: args{
				opts: Options{
					EnableFile: true,
					FilePath:   "",
				},
			},
			wantWriter: true,
			wantCloser: true,
			wantErr:    false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotW, gotCloser, err := setupWriter(tt.args.opts)
			if (err != nil) != tt.wantErr {
				t.Errorf("setupWriter() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if (gotW != nil) != tt.wantWriter {
				t.Errorf("setupWriter() gotW = %v, want non-nil = %v", gotW, tt.wantWriter)
			}
			if (gotCloser != nil) != tt.wantCloser {
				t.Errorf("setupWriter() gotCloser = %v, want non-nil = %v", gotCloser, tt.wantCloser)
			}
			if gotCloser != nil {
				_ = gotCloser.Close()
			}
		})
	}

	// Clean up default ./log if created
	_ = os.RemoveAll("./log")
}

func Test_validateOptions(t *testing.T) {
	tests := []struct {
		name        string
		input       Options
		wantFormat  Format
		wantMinLvl  Level
	}{
		{
			name:        "Empty options gets defaults",
			input:       Options{},
			wantFormat:  FormatJSON,
			wantMinLvl:  DEBUG,
		},
		{
			name: "Preserve valid text format and debug level",
			input: Options{
				Format:   FormatText,
				MinLevel: DEBUG,
			},
			wantFormat: FormatText,
			wantMinLvl: DEBUG,
		},
		{
			name: "Invalid format and out-of-range level gets reset",
			input: Options{
				Format:   "XML",
				MinLevel: Level(99),
			},
			wantFormat: FormatJSON,
			wantMinLvl: INFO,
		},
		{
			name: "Negative level gets reset to INFO",
			input: Options{
				Format:   FormatJSON,
				MinLevel: Level(-5),
			},
			wantFormat: FormatJSON,
			wantMinLvl: INFO,
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
		})
	}
}
