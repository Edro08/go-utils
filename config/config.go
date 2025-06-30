package config

// IConfig
// ------------------------------------------------------------------------------------------------
type IConfig interface {
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
	// Set(key string, value interface{}) error
	// Get(keys string) interface{}
	// GetKeys(keys string) []string
	// GetNestedConfig(keys string) IConfig
}

// ------------------------------------------------------------------------------------------------
// Implementation Methods
// ------------------------------------------------------------------------------------------------

type ValueType string

const (
	MapInterface ValueType = "map"
	MapString    ValueType = "mapString"
	MapInt       ValueType = "mapInt"
	MapFloat     ValueType = "mapFloat"
	MapBool      ValueType = "mapBool"

	SliceInterface ValueType = "slice"
	SliceString    ValueType = "sliceString"
	SliceInt       ValueType = "sliceInt"
	SliceFloat     ValueType = "sliceFloat"
	SliceBool      ValueType = "sliceBool"
)

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
	if v, ok := c.getMap(keys, MapInterface); ok {
		if result, ok := v.(map[string]interface{}); ok {
			return result
		}
	}
	return map[string]interface{}{}
}

func (c *Config) GetMapString(keys string) map[string]string {
	if v, ok := c.getMap(keys, MapString); ok {
		if result, ok := v.(map[string]string); ok {
			return result
		}
	}
	return map[string]string{}
}

func (c *Config) GetMapInt(keys string) map[string]int {
	if v, ok := c.getMap(keys, MapInt); ok {
		if result, ok := v.(map[string]int); ok {
			return result
		}
	}
	return map[string]int{}
}

func (c *Config) GetMapFloat(keys string) map[string]float64 {
	if v, ok := c.getMap(keys, MapFloat); ok {
		if result, ok := v.(map[string]float64); ok {
			return result
		}
	}
	return map[string]float64{}
}

func (c *Config) GetMapBool(keys string) map[string]bool {
	if v, ok := c.getMap(keys, MapBool); ok {
		if result, ok := v.(map[string]bool); ok {
			return result
		}
	}
	return map[string]bool{}
}

func (c *Config) GetSlice(keys string) []interface{} {
	if v, ok := c.getSlice(keys, SliceInterface); ok {
		if result, ok := v.([]interface{}); ok {
			return result
		}
	}
	return []interface{}{}
}

func (c *Config) GetSliceString(keys string) []string {
	if v, ok := c.getSlice(keys, SliceString); ok {
		if result, ok := v.([]string); ok {
			return result
		}
	}
	return []string{}
}

func (c *Config) GetSliceInt(keys string) []int {
	if v, ok := c.getSlice(keys, SliceInt); ok {
		if result, ok := v.([]int); ok {
			return result
		}
	}
	return []int{}
}

func (c *Config) GetSliceFloat(keys string) []float64 {
	if v, ok := c.getSlice(keys, SliceFloat); ok {
		if result, ok := v.([]float64); ok {
			return result
		}
	}
	return []float64{}
}

func (c *Config) GetSliceBool(keys string) []bool {
	if v, ok := c.getSlice(keys, SliceBool); ok {
		if result, ok := v.([]bool); ok {
			return result
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
	case MapInterface, MapString, MapInt, MapFloat, MapBool:
		_, ok := c.getMap(key, valueType)
		return ok
	case SliceInterface, SliceString, SliceInt, SliceFloat, SliceBool:
		_, ok := c.getSlice(key, valueType)
		return ok
	default:
		return false
	}
}
