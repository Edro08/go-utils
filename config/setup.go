package config

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"sync"

	"gopkg.in/yaml.v3"
)

var (
	ErrKeyEmpty      = errors.New("key cannot be empty")
	ErrReadFile      = errors.New("failed to read config file")
	ErrMarshalStruct = errors.New("failed to marshal struct")
	ErrParseYAML     = errors.New("failed to parse YAML")
	ErrMapKeyType    = errors.New("map key must be string")
	ErrInvalidTarget = errors.New("unmarshal target must be a non-nil pointer")
	ErrKeyNotFound   = errors.New("configuration key not found")
)

const (
	defaultSeparator = "."
)

// Options configura el comportamiento del gestor de configuración.
type Options struct {
	Separator    string
	ExpandEnv    bool
	AutomaticEnv bool
	EnvPrefix    string
}

type Config struct {
	data map[string]interface{}
	opts Options
	mu   sync.RWMutex
}

func New(opts Options) *Config {
	if opts.Separator == "" {
		opts.Separator = defaultSeparator
	}
	return &Config{
		data: make(map[string]interface{}),
		opts: opts,
	}
}

// LoadFile lee un archivo YAML desde disco y fusiona su contenido en la configuración actual.
// Si Options.ExpandEnv es true, expande expresiones ${VAR}, ${VAR:-default} y ${VAR-default}.
func (c *Config) LoadFile(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrReadFile, err)
	}

	return c.LoadBytes(data)
}

// LoadBytes parsea contenido YAML en memoria y lo fusiona profundamente con la configuración actual.
// Si Options.ExpandEnv es true, expande expresiones ${VAR}, ${VAR:-default} y ${VAR-default}.
func (c *Config) LoadBytes(data []byte) error {
	if c.opts.ExpandEnv && len(data) > 0 {
		data = []byte(expandEnvWithDefaults(string(data)))
	}

	m, err := unmarshalToMap(data)
	if err != nil {
		return err
	}
	if len(m) == 0 {
		return nil
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	if c.data == nil {
		c.data = make(map[string]interface{})
	}
	deepMerge(c.data, m)
	return nil
}

func (c *Config) LoadStruct(s interface{}) error {
	data, err := yaml.Marshal(s)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrMarshalStruct, err)
	}

	m, err := unmarshalToMap(data)
	if err != nil {
		return err
	}
	if len(m) == 0 {
		return nil
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	if c.data == nil {
		c.data = make(map[string]interface{})
	}
	deepMerge(c.data, m)
	return nil
}

func deepMerge(dst, src map[string]interface{}) {
	for k, srcVal := range src {
		if dstVal, exists := dst[k]; exists {
			srcMap, srcIsMap := srcVal.(map[string]interface{})
			dstMap, dstIsMap := dstVal.(map[string]interface{})
			if srcIsMap && dstIsMap {
				deepMerge(dstMap, srcMap)
				continue
			}
		}
		dst[k] = cloneValue(srcVal)
	}
}

func unmarshalToMap(data []byte) (map[string]interface{}, error) {
	var m map[string]interface{}
	if err := yaml.Unmarshal(data, &m); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrParseYAML, err)
	}
	if m == nil {
		return nil, nil
	}

	normalized, err := normalizeValue(m)
	if err != nil {
		return nil, err
	}
	nm, _ := normalized.(map[string]interface{})
	return nm, nil
}

// expandEnvWithDefaults expande únicamente expresiones con llaves ${VAR}, ${VAR:-default} y ${VAR-default},
// permitiendo escapar literales con $${VAR} sin alterar hashes o cadenas con '$' sueltos (ej. bcrypt).
func expandEnvWithDefaults(input string) string {
	var b strings.Builder
	b.Grow(len(input))

	for i := 0; i < len(input); {
		// Escape $${...} -> ${...}
		if i+2 < len(input) && input[i] == '$' && input[i+1] == '$' && input[i+2] == '{' {
			b.WriteByte('$')
			i += 2
			continue
		}

		if i+1 < len(input) && input[i] == '$' && input[i+1] == '{' {
			end := strings.IndexByte(input[i+2:], '}')
			if end != -1 {
				expr := input[i+2 : i+2+end]
				b.WriteString(resolveEnvExpr(expr))
				i += 2 + end + 1
				continue
			}
		}

		b.WriteByte(input[i])
		i++
	}

	return b.String()
}

func resolveEnvExpr(expr string) string {
	if idx := strings.Index(expr, ":-"); idx > 0 {
		name := expr[:idx]
		def := expr[idx+2:]
		if val, ok := os.LookupEnv(name); ok && val != "" {
			return val
		}
		return def
	}
	if idx := strings.IndexByte(expr, '-'); idx > 0 {
		name := expr[:idx]
		def := expr[idx+1:]
		if val, ok := os.LookupEnv(name); ok {
			return val
		}
		return def
	}
	if expr == "" {
		return ""
	}
	return os.Getenv(expr)
}
