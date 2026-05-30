package config

import "errors"

var (
	ErrKeyEmpty      = errors.New("key cannot be empty")
	ErrReadFile      = errors.New("failed to read config file")
	ErrMarshalStruct = errors.New("failed to marshal struct")
	ErrParseYAML     = errors.New("failed to parse YAML")
)
