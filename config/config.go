package config

import "fmt"

// IConfig
// ------------------------------------------------------------------------------------------------
type IConfig interface {
	GetString(keys string) string
	GetInt(keys string) int
	GetBool(keys string) bool
	GetMapInterface(keys string) map[string]interface{}
	GetMapString(keys string) map[string]string
	GetMapInt(keys string) map[string]int
	//GetSliceString(keys string) []string
	//GetSliceInt(keys string) []int
	//GetSliceBool(keys string) []bool
}

// ------------------------------------------------------------------------------------------------
// Implementation Methods
// ------------------------------------------------------------------------------------------------
const (
	mapStringName    = "mapString"
	mapIntName       = "mapInt"
	mapInterfaceName = "mapInterface"
)

func (c *Config) GetString(keys string) string {
	value, found := c.getValue(keys)
	if !found {
		return ""
	}
	return fmt.Sprintf("%v", value)
}

func (c *Config) GetInt(keys string) int {
	value, _ := c.getValue(keys)
	if intValue, ok := value.(int); ok {
		return intValue
	}
	return 0
}

func (c *Config) GetBool(keys string) bool {
	value, _ := c.getValue(keys)
	if boolValue, ok := value.(bool); ok {
		return boolValue
	}
	return false
}

func (c *Config) GetMapInterface(keys string) map[string]interface{} {
	value, _ := c.getMap(keys, mapInterfaceName)
	if value, ok := value.(map[string]interface{}); ok {
		return value
	}
	return nil
}

func (c *Config) GetMapString(keys string) map[string]string {
	value, _ := c.getMap(keys, mapStringName)
	if value, ok := value.(map[string]string); ok {
		return value
	}
	return nil
}

func (c *Config) GetMapInt(keys string) map[string]int {
	value, _ := c.getMap(keys, mapIntName)
	if value, ok := value.(map[string]int); ok {
		return value
	}
	return nil
}
