package config

import (
	"strings"
)

// getValue obtiene un valor dentro de la configuración basado en la clave proporcionada.
// La clave puede ser una ruta anidada separada por el separador configurado (por defecto ".").
// La función navega el mapa anidado y devuelve el valor encontrado y true si existe,
// o nil y false si la clave no existe o no se puede resolver completamente.
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

// getMap obtiene un mapa (map[string]interface{}) desde la configuración con la clave indicada.
// Devuelve el mapa y true si la clave existe y el valor es un mapa,
// o nil y false si no existe o el valor no es un mapa.
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

// getSlice obtiene un slice ([]interface{}) desde la configuración con la clave indicada.
// Devuelve el slice y true si la clave existe y el valor es un slice,
// o nil y false si no existe o el valor no es un slice.
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

// set asigna un valor en la configuración bajo la clave especificada.
// Si la ruta contiene claves anidadas, crea mapas intermedios según sea necesario para alojar el valor.
// Retorna true si la asignación fue exitosa, false en caso contrario (por ejemplo, clave vacía).
func (c *Config) set(key string, value interface{}) bool {
	keys := strings.Split(key, c.opts.Separator)
	if len(keys) == 0 {
		return false
	}

	cm := c.data
	for i := 0; i < len(keys)-1; i++ {
		k := keys[i]
		if next, ok := cm[k]; ok {
			if m, ok := next.(map[string]interface{}); ok {
				cm = m
			} else {
				// Si la clave existe pero no es un mapa, se sobrescribe con un nuevo mapa
				nm := make(map[string]interface{})
				cm[k] = nm
				cm = nm
			}
		} else {
			// Si la clave no existe, se crea un nuevo mapa anidado
			nm := make(map[string]interface{})
			cm[k] = nm
			cm = nm
		}
	}

	// Finalmente, se asigna el valor en la última clave
	cm[keys[len(keys)-1]] = value
	return true
}
