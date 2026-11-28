package config

import (
	"fmt"
	"reflect"
	"time"

	"gopkg.in/yaml.v3"
)

// IConfig define el contrato para lectura y escritura segura de configuraciones jerárquicas.
// ------------------------------------------------------------------------------------------------
type IConfig interface {
	Get(keys string) interface{}
	GetString(keys string) string
	GetInt(keys string) int
	GetFloat(keys string) float64
	GetBool(keys string) bool
	GetDuration(keys string) time.Duration

	GetStringOr(keys string, defaultVal string) string
	GetIntOr(keys string, defaultVal int) int
	GetFloatOr(keys string, defaultVal float64) float64
	GetBoolOr(keys string, defaultVal bool) bool
	GetDurationOr(keys string, defaultVal time.Duration) time.Duration

	GetMap(keys string) map[string]interface{}
	GetMapString(keys string) map[string]string
	GetMapInt(keys string) map[string]int
	GetMapFloat(keys string) map[string]float64
	GetMapBool(keys string) map[string]bool

	GetSlice(keys string) []interface{}
	GetSliceString(keys string) []string
	GetSliceInt(keys string) []int
	GetSliceFloat(keys string) []float64
	GetSliceBool(keys string) []bool

	Has(keys string) bool
	HasKey(keys string, valueType ValueType) bool
	GetKeys(keys string) []string
	Set(key string, value interface{}) error
	Unmarshal(keys string, out interface{}) error
}

// ------------------------------------------------------------------------------------------------
// Implementation Methods
// ------------------------------------------------------------------------------------------------

type ValueType string

const (
	Map   ValueType = "Map"
	Slice ValueType = "Slice"
)

// Get devuelve el valor asociado a la clave especificada como interface{}.
// Si la clave no existe, retorna nil.
func (c *Config) Get(keys string) interface{} {
	c.mu.RLock()
	defer c.mu.RUnlock()
	v, _ := c.getValue(keys)
	return v
}

// GetString devuelve el valor de la clave como string.
// Si no existe o no es convertible, retorna una cadena vacía.
func (c *Config) GetString(keys string) string {
	return c.GetStringOr(keys, "")
}

// GetStringOr devuelve el valor de la clave como string o defaultVal si no existe o no es convertible.
func (c *Config) GetStringOr(keys string, defaultVal string) string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if v, ok := c.findValue(keys); ok {
		if s, ok := parseString(v); ok {
			return s
		}
	}
	return defaultVal
}

// GetInt devuelve el valor de la clave como int.
// Si no existe o no es convertible, retorna 0.
func (c *Config) GetInt(keys string) int {
	return c.GetIntOr(keys, 0)
}

// GetIntOr devuelve el valor de la clave como int o defaultVal si no existe o no es convertible.
func (c *Config) GetIntOr(keys string, defaultVal int) int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if v, ok := c.findValue(keys); ok {
		if i, ok := parseInt(v); ok {
			return i
		}
	}
	return defaultVal
}

// GetFloat devuelve el valor de la clave como float64.
// Si no existe o no es convertible, retorna 0.0.
func (c *Config) GetFloat(keys string) float64 {
	return c.GetFloatOr(keys, 0.0)
}

// GetFloatOr devuelve el valor de la clave como float64 o defaultVal si no existe o no es convertible.
func (c *Config) GetFloatOr(keys string, defaultVal float64) float64 {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if v, ok := c.findValue(keys); ok {
		if f, ok := parseFloat64(v); ok {
			return f
		}
	}
	return defaultVal
}

// GetBool devuelve el valor de la clave como bool.
// Si no existe o no es convertible, retorna false.
func (c *Config) GetBool(keys string) bool {
	return c.GetBoolOr(keys, false)
}

// GetBoolOr devuelve el valor de la clave como bool o defaultVal si no existe o no es convertible.
// Distingue correctamente un valor explícito false en la configuración de una clave inexistente.
func (c *Config) GetBoolOr(keys string, defaultVal bool) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if v, ok := c.findValue(keys); ok {
		if b, ok := parseBool(v); ok {
			return b
		}
	}
	return defaultVal
}

// GetDuration devuelve el valor de la clave como time.Duration (ej. "5s", "250ms").
// Si no existe o no es válido, retorna 0.
func (c *Config) GetDuration(keys string) time.Duration {
	return c.GetDurationOr(keys, 0)
}

// GetDurationOr devuelve el valor de la clave como time.Duration o defaultVal si no existe o no es válido.
func (c *Config) GetDurationOr(keys string, defaultVal time.Duration) time.Duration {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if v, ok := c.findValue(keys); ok {
		if d, ok := parseDuration(v); ok {
			return d
		}
	}
	return defaultVal
}

// GetMap devuelve el valor de la clave como un map[string]interface{} clonado de forma segura.
// Si la clave no existe o no es un mapa, retorna un mapa vacío.
func (c *Config) GetMap(keys string) map[string]interface{} {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if m, ok := c.getMap(keys); ok {
		return m
	}
	return map[string]interface{}{}
}

// GetMapString convierte y retorna un map[string]string.
// Si la clave no existe o no es un mapa, retorna un mapa vacío.
func (c *Config) GetMapString(keys string) map[string]string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if m, ok := c.getMap(keys); ok {
		if r, ok := convertToStringMap(m); ok {
			return r
		}
	}
	return map[string]string{}
}

// GetMapInt convierte y retorna un map[string]int.
// Si la clave no existe o no es un mapa, retorna un mapa vacío.
func (c *Config) GetMapInt(keys string) map[string]int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if m, ok := c.getMap(keys); ok {
		if r, ok := convertToIntMap(m); ok {
			return r
		}
	}
	return map[string]int{}
}

// GetMapFloat convierte y retorna un map[string]float64.
// Si la clave no existe o no es un mapa, retorna un mapa vacío.
func (c *Config) GetMapFloat(keys string) map[string]float64 {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if m, ok := c.getMap(keys); ok {
		if r, ok := convertToFloatMap(m); ok {
			return r
		}
	}
	return map[string]float64{}
}

// GetMapBool convierte y retorna un map[string]bool.
// Si la clave no existe o no es un mapa, retorna un mapa vacío.
func (c *Config) GetMapBool(keys string) map[string]bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if m, ok := c.getMap(keys); ok {
		if r, ok := convertToBoolMap(m); ok {
			return r
		}
	}
	return map[string]bool{}
}

// GetSlice devuelve el valor de la clave como un []interface{} clonado de forma segura.
// Si la clave no existe o no es un slice, retorna un slice vacío.
func (c *Config) GetSlice(keys string) []interface{} {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if s, ok := c.getSlice(keys); ok {
		return s
	}
	return []interface{}{}
}

// GetSliceString convierte y retorna un []string.
// Si la clave no existe o no es un slice, retorna un slice vacío.
func (c *Config) GetSliceString(keys string) []string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if s, ok := c.getSlice(keys); ok {
		if r, ok := convertToStringSlice(s); ok {
			return r
		}
	}
	return []string{}
}

// GetSliceInt convierte y retorna un []int.
// Si la clave no existe o no es un slice, retorna un slice vacío.
func (c *Config) GetSliceInt(keys string) []int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if s, ok := c.getSlice(keys); ok {
		if r, ok := convertToIntSlice(s); ok {
			return r
		}
	}
	return []int{}
}

// GetSliceFloat convierte y retorna un []float64.
// Si la clave no existe o no es un slice, retorna un slice vacío.
func (c *Config) GetSliceFloat(keys string) []float64 {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if s, ok := c.getSlice(keys); ok {
		if r, ok := convertToFloatSlice(s); ok {
			return r
		}
	}
	return []float64{}
}

// GetSliceBool convierte y retorna un []bool.
// Si la clave no existe o no es un slice, retorna un slice vacío.
func (c *Config) GetSliceBool(keys string) []bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if s, ok := c.getSlice(keys); ok {
		if r, ok := convertToBoolSlice(s); ok {
			return r
		}
	}
	return []bool{}
}

// Has indica si una clave existe en la configuración (incluyendo AutomaticEnv si está activo).
func (c *Config) Has(key string) bool {
	return c.HasKey(key, "")
}

// HasKey indica si una clave existe en la configuración.
// También permite validar si es del tipo esperado (mapa o slice).
func (c *Config) HasKey(key string, valueType ValueType) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if valueType == "" {
		_, ok := c.findValue(key)
		return ok
	}

	switch valueType {
	case Map:
		_, ok := c.getMap(key)
		return ok
	case Slice:
		_, ok := c.getSlice(key)
		return ok
	default:
		return false
	}
}

// GetKeys retorna todas las claves del mapa asociado a la clave dada.
// Si keys es vacío (""), retorna las claves raíz de la configuración.
// Si la clave no existe o no es un mapa, retorna un slice vacío.
func (c *Config) GetKeys(keys string) []string {
	c.mu.RLock()
	defer c.mu.RUnlock()

	var m map[string]interface{}
	if keys == "" {
		m = c.data
	} else {
		var ok bool
		m, ok = c.getMap(keys)
		if !ok {
			return []string{}
		}
	}

	res := make([]string, 0, len(m))
	for k := range m {
		res = append(res, k)
	}
	return res
}

// Set establece o actualiza un valor dentro de la configuración utilizando una clave jerárquica.
func (c *Config) Set(key string, value interface{}) error {
	return c.set(key, value)
}

// Unmarshal decodifica la sección indicada por keys (o toda la configuración si keys es "")
// dentro del puntero a estructura o mapa out.
func (c *Config) Unmarshal(keys string, out interface{}) error {
	if out == nil {
		return ErrInvalidTarget
	}
	rv := reflect.ValueOf(out)
	if rv.Kind() != reflect.Pointer || rv.IsNil() {
		return ErrInvalidTarget
	}

	c.mu.RLock()
	var target interface{}
	if keys == "" {
		target = cloneValue(c.data)
	} else {
		val, ok := c.getValue(keys)
		if !ok {
			c.mu.RUnlock()
			return fmt.Errorf("%w: %s", ErrKeyNotFound, keys)
		}
		target = val
	}
	if c.opts.AutomaticEnv {
		if tm, ok := target.(map[string]interface{}); ok {
			c.applyEnvOverridesToMap(keys, tm)
		}
	}
	c.mu.RUnlock()

	raw, err := yaml.Marshal(target)
	if err != nil {
		return fmt.Errorf("%w: %s", ErrMarshalStruct, err)
	}
	if err := yaml.Unmarshal(raw, out); err != nil {
		return fmt.Errorf("%w: %s", ErrParseYAML, err)
	}
	return nil
}
