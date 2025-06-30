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
	// Set(key string, value interface{}) error
}

// ------------------------------------------------------------------------------------------------
// Implementation Methods
// ------------------------------------------------------------------------------------------------

type ValueType string

const (
	Map   ValueType = "Map"
	Slice ValueType = "Slice"
)

func (c *Config) Get(keys string) interface{} {
	if v, ok := c.getValue(keys); ok {
		return v
	}
	return nil

}

func (c *Config) GetString(keys string) string {
	if v, ok := c.getValue(keys); ok {
		return toString(v)
	}
	return ""
}

func (c *Config) GetInt(keys string) int {
	if v, ok := c.getValue(keys); ok {
		return toInt(v)
	}
	return 0
}

func (c *Config) GetBool(keys string) bool {
	if v, ok := c.getValue(keys); ok {
		return toBool(v)
	}
	return false
}

func (c *Config) GetFloat(keys string) float64 {
	if v, ok := c.getValue(keys); ok {
		return toFloat64(v)
	}
	return 0.0
}
func (c *Config) GetMap(keys string) map[string]interface{} {
	if v, ok := c.getMap(keys); ok {
		return v
	}
	return map[string]interface{}{}
}

func (c *Config) GetMapString(keys string) map[string]string {
	if v, ok := c.getMap(keys); ok {
		if r, ok := convertToStringMap(v); ok {
			return r
		}
	}
	return map[string]string{}
}

func (c *Config) GetMapInt(keys string) map[string]int {
	if v, ok := c.getMap(keys); ok {
		if r, ok := convertToIntMap(v); ok {
			return r
		}
	}
	return map[string]int{}
}

func (c *Config) GetMapFloat(keys string) map[string]float64 {
	if v, ok := c.getMap(keys); ok {
		if r, ok := convertToFloatMap(v); ok {
			return r
		}
	}
	return map[string]float64{}
}

func (c *Config) GetMapBool(keys string) map[string]bool {
	if v, ok := c.getMap(keys); ok {
		if r, ok := convertToBoolMap(v); ok {
			return r
		}
	}
	return map[string]bool{}
}

func (c *Config) GetSlice(keys string) []interface{} {
	if v, ok := c.getSlice(keys); ok {
		return v
	}
	return []interface{}{}
}

func (c *Config) GetSliceString(keys string) []string {
	if v, ok := c.getSlice(keys); ok {
		if r, ok := convertToStringSlice(v); ok {
			return r
		}
	}
	return []string{}
}

func (c *Config) GetSliceInt(keys string) []int {
	if v, ok := c.getSlice(keys); ok {
		if r, ok := convertToIntSlice(v); ok {
			return r
		}
	}
	return []int{}
}

func (c *Config) GetSliceFloat(keys string) []float64 {
	if v, ok := c.getSlice(keys); ok {
		if r, ok := convertToFloatSlice(v); ok {
			return r
		}
	}
	return []float64{}
}

func (c *Config) GetSliceBool(keys string) []bool {
	if v, ok := c.getSlice(keys); ok {
		if r, ok := convertToBoolSlice(v); ok {
			return r
		}
	}
	return []bool{}
}

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

func (c *Config) GetNestedConfig(keys string) IConfig {
	v, ok := c.getValue(keys)
	if !ok {
		return &Config{data: map[string]interface{}{}}
	}

	sm, ok := v.(map[string]interface{})
	if !ok {
		return &Config{data: map[string]interface{}{}}
	}

	return &Config{data: sm}
}
