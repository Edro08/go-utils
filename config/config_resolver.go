package config

import (
	"fmt"
	"strings"
)

// getValue retrieves a value from the configuration data based on the provided key.
func (c *Config) getValue(key string) (interface{}, bool) {
	keyParts := strings.Split(key, ".")
	currentMap := c.data
	for _, part := range keyParts {
		value, ok := currentMap[part]
		if !ok {
			return nil, false
		}
		if next, ok := value.(map[string]interface{}); ok {
			currentMap = next
		} else {
			return value, true
		}
	}
	return currentMap, true
}

// getMap retrieves a map from the configuration data based on the provided key and type.
func (c *Config) getMap(key string, typeMap string) (interface{}, bool) {
	keyParts := strings.Split(key, ".")
	currentMap := c.data
	for _, part := range keyParts {
		value, ok := currentMap[part]
		if !ok {
			return nil, false
		}
		if nextMap, ok := value.(map[string]interface{}); ok {
			currentMap = nextMap
		}
	}

	switch typeMap {
	case mapInterfaceName:
		return currentMap, true
	case mapStringName:
		return convertToStringMap(currentMap)
	case mapIntName:
		return convertToIntMap(currentMap)
	default:
		return nil, false
	}
}

// convertToStringMap converts a map to map[string]string.
func convertToStringMap(input map[string]interface{}) (map[string]string, bool) {
	data := make(map[string]string)
	for k, v := range input {
		data[k] = fmt.Sprintf("%v", v)
	}
	return data, true
}

// convertToIntMap converts a map to map[string]int.
func convertToIntMap(input map[string]interface{}) (map[string]int, bool) {
	data := make(map[string]int)
	for k, v := range input {
		if intValue, ok := v.(int); ok {
			data[k] = intValue
		}
	}
	return data, true
}
