package config

import (
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestNew(t *testing.T) {
	type args struct {
		opts Options
	}
	tests := []struct {
		name          string
		args          args
		wantSeparator string
	}{
		{
			name: "Default separator",
			args: args{
				opts: Options{},
			},
			wantSeparator: defaultSeparator,
		},
		{
			name: "Custom separator",
			args: args{
				opts: Options{Separator: "/"},
			},
			wantSeparator: "/",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := New(tt.args.opts)
			if got == nil {
				t.Fatal("New() returned nil")
			}
			if got.opts.Separator != tt.wantSeparator {
				t.Errorf("New().opts.Separator = %v, want %v", got.opts.Separator, tt.wantSeparator)
			}
			if got.data == nil {
				t.Error("New().data is nil, expected initialized map")
			}
		})
	}
}

func TestConfig_LoadFile(t *testing.T) {
	tempDir := t.TempDir()

	validYAMLPath := filepath.Join(tempDir, "valid.yaml")
	validContent := []byte("app:\n  name: my-app\n  port: 8080\n")
	if err := os.WriteFile(validYAMLPath, validContent, 0600); err != nil {
		t.Fatalf("failed to create valid yaml: %v", err)
	}

	invalidYAMLPath := filepath.Join(tempDir, "invalid.yaml")
	invalidContent := []byte("app:\n  port: [invalid yaml\n")
	if err := os.WriteFile(invalidYAMLPath, invalidContent, 0600); err != nil {
		t.Fatalf("failed to create invalid yaml: %v", err)
	}

	type fields struct {
		data map[string]interface{}
		opts Options
	}
	type args struct {
		path string
	}
	tests := []struct {
		name      string
		fields    fields
		args      args
		wantErr   bool
		targetErr error
		wantData  map[string]interface{}
	}{
		{
			name: "Successfully load valid YAML",
			fields: fields{
				data: make(map[string]interface{}),
				opts: Options{Separator: "."},
			},
			args:    args{path: validYAMLPath},
			wantErr: false,
			wantData: map[string]interface{}{
				"app": map[string]interface{}{
					"name": "my-app",
					"port": 8080,
				},
			},
		},
		{
			name: "File does not exist",
			fields: fields{
				data: make(map[string]interface{}),
				opts: Options{Separator: "."},
			},
			args:      args{path: filepath.Join(tempDir, "non_existent.yaml")},
			wantErr:   true,
			targetErr: ErrReadFile,
		},
		{
			name: "Invalid YAML syntax",
			fields: fields{
				data: make(map[string]interface{}),
				opts: Options{Separator: "."},
			},
			args:      args{path: invalidYAMLPath},
			wantErr:   true,
			targetErr: ErrParseYAML,
		},
		{
			name: "Deep merge with existing data",
			fields: fields{
				data: map[string]interface{}{
					"app": map[string]interface{}{
						"env": "production",
					},
					"db": "postgres",
				},
				opts: Options{Separator: "."},
			},
			args:    args{path: validYAMLPath},
			wantErr: false,
			wantData: map[string]interface{}{
				"app": map[string]interface{}{
					"name": "my-app",
					"port": 8080,
					"env":  "production",
				},
				"db": "postgres",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Config{
				data: tt.fields.data,
				opts: tt.fields.opts,
			}
			err := c.LoadFile(tt.args.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("LoadFile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.targetErr != nil && !errors.Is(err, tt.targetErr) {
				t.Errorf("LoadFile() error = %v, targetErr %v", err, tt.targetErr)
			}
			if tt.wantData != nil && !reflect.DeepEqual(c.data, tt.wantData) {
				t.Errorf("LoadFile() data = %v, want %v", c.data, tt.wantData)
			}
		})
	}
}

type badMarshalStruct struct{}

func (b badMarshalStruct) MarshalYAML() (interface{}, error) {
	return nil, errors.New("custom marshal error")
}

func TestConfig_LoadStruct(t *testing.T) {
	type ServerConfig struct {
		Host string `yaml:"host"`
		Port int    `yaml:"port"`
	}

	type AppConfig struct {
		Name   string       `yaml:"name"`
		Server ServerConfig `yaml:"server"`
	}

	type fields struct {
		data map[string]interface{}
		opts Options
	}
	type args struct {
		s interface{}
	}
	tests := []struct {
		name      string
		fields    fields
		args      args
		wantErr   bool
		targetErr error
		wantData  map[string]interface{}
	}{
		{
			name: "Load valid struct",
			fields: fields{
				data: make(map[string]interface{}),
				opts: Options{Separator: "."},
			},
			args: args{
				s: AppConfig{
					Name: "microservice",
					Server: ServerConfig{
						Host: "localhost",
						Port: 9000,
					},
				},
			},
			wantErr: false,
			wantData: map[string]interface{}{
				"name": "microservice",
				"server": map[string]interface{}{
					"host": "localhost",
					"port": 9000,
				},
			},
		},
		{
			name: "Load struct pointer with deep merge",
			fields: fields{
				data: map[string]interface{}{
					"server": map[string]interface{}{
						"timeout": 30,
					},
				},
				opts: Options{Separator: "."},
			},
			args: args{
				s: &AppConfig{
					Name: "microservice",
					Server: ServerConfig{
						Host: "127.0.0.1",
						Port: 8080,
					},
				},
			},
			wantErr: false,
			wantData: map[string]interface{}{
				"name": "microservice",
				"server": map[string]interface{}{
					"host":    "127.0.0.1",
					"port":    8080,
					"timeout": 30,
				},
			},
		},
		{
			name: "Marshal error with unsupported type",
			fields: fields{
				data: make(map[string]interface{}),
				opts: Options{Separator: "."},
			},
			args: args{
				s: badMarshalStruct{},
			},
			wantErr:   true,
			targetErr: ErrMarshalStruct,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Config{
				data: tt.fields.data,
				opts: tt.fields.opts,
			}
			err := c.LoadStruct(tt.args.s)
			if (err != nil) != tt.wantErr {
				t.Errorf("LoadStruct() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.targetErr != nil && !errors.Is(err, tt.targetErr) {
				t.Errorf("LoadStruct() error = %v, targetErr %v", err, tt.targetErr)
			}
			if tt.wantData != nil && !reflect.DeepEqual(c.data, tt.wantData) {
				t.Errorf("LoadStruct() data = %v, want %v", c.data, tt.wantData)
			}
		})
	}
}
