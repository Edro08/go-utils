package config

// IConfig
// ------------------------------------------------------------------------------------------------
type IConfig interface {
	Get(keys string) interface{}
	GetString(keys string) string
	GetInt(keys string) int
	GetFloat(keys string) float64
	GetBool(keys string) bool

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

	HasKey(keys string, valueType ValueType) bool
	GetKeys(keys string) []string
	GetNestedConfig(keys string) IConfig
	Set(key string, value interface{}) bool
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
	if v, ok := c.getValue(keys); ok {
		return v
	}
	return nil
}

// GetString devuelve el valor de la clave como string.
// Si no existe o no es convertible, retorna una cadena vacía.
func (c *Config) GetString(keys string) string {
	if v, ok := c.getValue(keys); ok {
		return toString(v)
	}
	return ""
}

// GetInt devuelve el valor de la clave como int.
// Si no existe o no es convertible, retorna 0.
func (c *Config) GetInt(keys string) int {
	if v, ok := c.getValue(keys); ok {
		return toInt(v)
	}
	return 0
}

// GetFloat devuelve el valor de la clave como float64.
// Si no existe o no es convertible, retorna 0.0.
func (c *Config) GetFloat(keys string) float64 {
	if v, ok := c.getValue(keys); ok {
		return toFloat64(v)
	}
	return 0.0
}

// GetBool devuelve el valor de la clave como bool.
// Si no existe o no es convertible, retorna false.
func (c *Config) GetBool(keys string) bool {
	if v, ok := c.getValue(keys); ok {
		return toBool(v)
	}
	return false
}

// GetMap devuelve un mapa del valor asociado a la clave.
// Si no existe o no es un map[string]interface{}, retorna un mapa vacío.
func (c *Config) GetMap(keys string) map[string]interface{} {
	if m, ok := c.getMap(keys); ok {
		return m
	}
	return map[string]interface{}{}
}

// GetMapString convierte y retorna un map[string]string.
// Si la clave no existe o no es convertible, retorna un mapa vacío.
func (c *Config) GetMapString(keys string) map[string]string {
	if m, ok := c.getMap(keys); ok {
		if r, ok := convertToStringMap(m); ok {
			return r
		}
	}
	return map[string]string{}
}

// GetMapInt convierte y retorna un map[string]int.
// Si la clave no existe o no es convertible, retorna un mapa vacío.
func (c *Config) GetMapInt(keys string) map[string]int {
	if m, ok := c.getMap(keys); ok {
		if r, ok := convertToIntMap(m); ok {
			return r
		}
	}
	return map[string]int{}
}

// GetMapFloat convierte y retorna un map[string]float64.
// Si la clave no existe o no es convertible, retorna un mapa vacío.
func (c *Config) GetMapFloat(keys string) map[string]float64 {
	if m, ok := c.getMap(keys); ok {
		if r, ok := convertToFloatMap(m); ok {
			return r
		}
	}
	return map[string]float64{}
}

// GetMapBool convierte y retorna un map[string]bool.
// Si la clave no existe o no es convertible, retorna un mapa vacío.
func (c *Config) GetMapBool(keys string) map[string]bool {
	if m, ok := c.getMap(keys); ok {
		if r, ok := convertToBoolMap(m); ok {
			return r
		}
	}
	return map[string]bool{}
}

// GetSlice devuelve un slice genérico []interface{}.
// Si la clave no existe o no es un slice, retorna un slice vacío.
func (c *Config) GetSlice(keys string) []interface{} {
	if s, ok := c.getSlice(keys); ok {
		return s
	}
	return []interface{}{}
}

// GetSliceString convierte y retorna un []string.
// Si la clave no existe o no es convertible, retorna un slice vacío.
func (c *Config) GetSliceString(keys string) []string {
	if s, ok := c.getSlice(keys); ok {
		if r, ok := convertToStringSlice(s); ok {
			return r
		}
	}
	return []string{}
}

// GetSliceInt convierte y retorna un []int.
// Si la clave no existe o no es convertible, retorna un slice vacío.
func (c *Config) GetSliceInt(keys string) []int {
	if s, ok := c.getSlice(keys); ok {
		if r, ok := convertToIntSlice(s); ok {
			return r
		}
	}
	return []int{}
}

// GetSliceFloat convierte y retorna un []float64.
// Si la clave no existe o no es convertible, retorna un slice vacío.
func (c *Config) GetSliceFloat(keys string) []float64 {
	if s, ok := c.getSlice(keys); ok {
		if r, ok := convertToFloatSlice(s); ok {
			return r
		}
	}
	return []float64{}
}

// GetSliceBool convierte y retorna un []bool.
// Si la clave no existe o no es convertible, retorna un slice vacío.
func (c *Config) GetSliceBool(keys string) []bool {
	if s, ok := c.getSlice(keys); ok {
		if r, ok := convertToBoolSlice(s); ok {
			return r
		}
	}
	return []bool{}
}

// HasKey indica si una clave existe en la configuración.
// También permite validar si es del tipo esperado (mapa o slice).
func (c *Config) HasKey(key string, valueType ValueType) bool {
	if valueType == "" {
		_, ok := c.getValue(key)
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
// Si la clave no existe o no es un mapa, retorna un slice vacío.
func (c *Config) GetKeys(keys string) []string {
	v, ok := c.getValue(keys)
	if !ok {
		return []string{}
	}

	if m, ok := v.(map[string]interface{}); ok {
		keys := make([]string, 0, len(m))
		for k := range m {
			keys = append(keys, k)
		}
		return keys
	}

	return []string{}
}

// GetNestedConfig retorna una nueva instancia de configuración basada en un submapa.
// Si la clave no es un mapa, se devuelve una configuración vacía.
func (c *Config) GetNestedConfig(keys string) IConfig {
	if m, ok := c.getMap(keys); ok {
		return &Config{data: m, opts: c.opts}
	}

	return &Config{data: map[string]interface{}{}, opts: c.opts}
}

// Set establece un valor en la ruta especificada, creando los niveles anidados necesarios.
// Retorna true si el valor fue asignado correctamente.
func (c *Config) Set(key string, value interface{}) bool {
	return c.set(key, value)
}
