package config

import (
	"strings"
)

// getValue retrieves a value from the configuration data based on the provided key.
func (c *Config) getValue(key string) (interface{}, bool) {
	keys := strings.Split(key, c.opts.Separator)
	current := c.data
	for i, k := range keys {
		val, ok := current[k]
		if !ok {
			return nil, false
		}

		if m, ok := val.(map[string]interface{}); ok && i < len(keys)-1 {
			current = m
		} else {
			return val, true
		}
	}

	return current, true
}

func (c *Config) getMap(key string) (map[string]interface{}, bool) {
	v, ok := c.getValue(key)
	if !ok {
		return nil, false
	}

	m, ok := v.(map[string]interface{})
	if !ok {
		return nil, false
	}

	return m, true
}

func (c *Config) getSlice(key string) ([]interface{}, bool) {
	v, ok := c.getValue(key)
	if !ok {
		return nil, false
	}

	s, ok := v.([]interface{})
	if !ok {
		return nil, false
	}

	return s, true
}
