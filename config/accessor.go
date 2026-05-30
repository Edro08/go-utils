package config

import (
	"strings"
)

// cloneValue realiza una copia profunda de cualquier valor.
// Previene que mapas y slices retornados o inyectados mantengan la referencia original.
func cloneValue(v interface{}) interface{} {
	switch val := v.(type) {
	case map[string]interface{}:
		cp := make(map[string]interface{}, len(val))
		for k, v2 := range val {
			cp[k] = cloneValue(v2)
		}
		return cp
	case []interface{}:
		cp := make([]interface{}, len(val))
		for i, v2 := range val {
			cp[i] = cloneValue(v2)
		}
		return cp
	default:
		// Los tipos primitivos (string, int, bool, etc.) en Go se pasan por valor.
		return val
	}
}
func (c *Config) getRawValue(key string) (interface{}, bool) {
	if key == "" {
		return nil, false
	}

	keys := strings.Split(key, c.opts.Separator)
	var current interface{} = c.data

	for _, k := range keys {
		m, ok := current.(map[string]interface{})
		if !ok {
			return nil, false
		}
		val, exists := m[k]
		if !exists {
			return nil, false
		}
		current = val
	}

	return current, true
}

// getValue usa getRawValue y clona el resultado genérico.
func (c *Config) getValue(key string) (interface{}, bool) {
	val, ok := c.getRawValue(key)
	if !ok {
		return nil, false
	}
	return cloneValue(val), true
}

// getRawMap devuelve la referencia original del mapa (para uso en conversores)
func (c *Config) getRawMap(key string) (map[string]interface{}, bool) {
	v, ok := c.getRawValue(key)
	if !ok {
		return nil, false
	}
	m, ok := v.(map[string]interface{})
	return m, ok
}

// getRawSlice devuelve la referencia original del slice (para uso en conversores)
func (c *Config) getRawSlice(key string) ([]interface{}, bool) {
	v, ok := c.getRawValue(key)
	if !ok {
		return nil, false
	}
	s, ok := v.([]interface{})
	return s, ok
}

func (c *Config) set(key string, value interface{}) error {
	c.mu.Lock()         // Bloqueo total: nadie puede leer ni escribir
	defer c.mu.Unlock() // Se libera automáticamente al terminar la función

	if key == "" {
		return ErrKeyEmpty
	}

	keys := strings.Split(key, c.opts.Separator)
	cm := c.data

	for i := 0; i < len(keys)-1; i++ {
		k := keys[i]
		next, exists := cm[k]

		if exists {
			if m, ok := next.(map[string]interface{}); ok {
				cm = m
				continue
			}
		}

		nm := make(map[string]interface{})
		cm[k] = nm
		cm = nm
	}

	cm[keys[len(keys)-1]] = cloneValue(value)
	return nil
}
