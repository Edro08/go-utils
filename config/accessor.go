package config

import (
	"strings"
)

// getValue retrieves a value from the configuration data based on the provided key.
func (c *Config) getValue(key string) (interface{}, bool) {
	keys := strings.Split(key, ".")
	current := c.data
	for i, k := range keys {
		val, ok := current[k]
		if !ok {
			return nil, false
		}
		// If it's a map and not the last key, go deeper
		if m, ok := val.(map[string]interface{}); ok && i < len(keys)-1 {
			current = m
		} else {
			return val, true
		}
	}
	// Return the map itself if last key resolves to a map
	return current, true
}

func (c *Config) getMap(key string, typeMap ValueType) (interface{}, bool) {
	// Obtener el valor directamente con getValue
	value, found := c.getValue(key)
	if !found {
		return nil, false
	}

	// Validar que sea un map[string]interface{}
	m, ok := value.(map[string]interface{})
	if !ok {
		return nil, false
	}

	// Delegar la conversión según el tipo
	switch typeMap {
	case MapInterface:
		return m, true
	case MapString:
		return convertToStringMap(m)
	case MapInt:
		return convertToIntMap(m)
	case MapFloat:
		return convertToFloatMap(m)
	case MapBool:
		return convertToBoolMap(m)
	default:
		return nil, false
	}
}

func (c *Config) getSlice(key string, typeSlice ValueType) (interface{}, bool) {
	// Obtener el valor directamente con getValue
	value, found := c.getValue(key)
	if !found {
		return nil, false
	}

	// Validar que sea un []interface{}
	rs, ok := value.([]interface{})
	if !ok {
		return nil, false
	}

	switch typeSlice {
	case SliceInterface:
		return rs, true
	case SliceString:
		return convertToStringSlice(rs)
	case SliceInt:
		return convertToIntSlice(rs)
	case SliceFloat:
		return convertToFloatSlice(rs)
	case SliceBool:
		return convertToBoolSlice(rs)
	default:
		return nil, false
	}
}
