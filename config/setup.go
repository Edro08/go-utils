package config

import (
	"fmt"
	"gopkg.in/yaml.v3"
	"os"
)

// ------------------------------------------------------------------------------------------------
// Struct Options
// ------------------------------------------------------------------------------------------------

const (
	defaultSeparator = "."
)

type Opts struct {
	File      string
	Separator string
}

// ------------------------------------------------------------------------------------------------
// NewConfig
// ------------------------------------------------------------------------------------------------

type Config struct {
	data map[string]interface{}
	opts Opts
}

func NewConfig(opts Opts) (*Config, error) {
	yamlFile, err := os.ReadFile(opts.File)
	if err != nil {
		return nil, fmt.Errorf("error leyendo archivo YAML: %w", err)
	}

	var data map[string]interface{}

	err = yaml.Unmarshal(yamlFile, &data)
	if err != nil {
		return nil, fmt.Errorf("error unmarshalling YAML: %w", err)
	}

	if opts.Separator == "" {
		opts.Separator = defaultSeparator
	}

	return &Config{data: data, opts: opts}, nil
}
