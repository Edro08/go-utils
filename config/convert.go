package config

import (
	"fmt"
	"strconv"
	"strings"
)

// convertToStringMap converts a map to map[string]string.
func convertToStringMap(input map[string]interface{}) (map[string]string, bool) {
	data := make(map[string]string)
	for k, v := range input {
		data[k] = toString(v)
	}
	return data, true
}

// convertToIntMap converts a map to map[string]int.
func convertToIntMap(input map[string]interface{}) (map[string]int, bool) {
	data := make(map[string]int)
	for k, v := range input {
		data[k] = toInt(v)
	}
	return data, true
}

func convertToFloatMap(input map[string]interface{}) (map[string]float64, bool) {
	data := make(map[string]float64)
	for k, v := range input {
		data[k] = toFloat64(v)
	}
	return data, true
}

func convertToBoolMap(input map[string]interface{}) (map[string]bool, bool) {
	data := make(map[string]bool)
	for k, v := range input {
		data[k] = toBool(v)
	}
	return data, true
}

// convertToStringSlice converts a slice to []string.
func convertToStringSlice(input []interface{}) ([]string, bool) {
	result := make([]string, 0, len(input))
	for _, v := range input {
		result = append(result, toString(v))
	}
	return result, true
}

// convertToIntSlice converts a slice to []int.
func convertToIntSlice(input []interface{}) ([]int, bool) {
	result := make([]int, 0, len(input))
	for _, v := range input {
		result = append(result, toInt(v))
	}
	return result, true
}

// convertToFloatSlice converts a slice to []float64.
func convertToFloatSlice(input []interface{}) ([]float64, bool) {
	result := make([]float64, 0, len(input))
	for _, v := range input {
		result = append(result, toFloat64(v))
	}
	return result, true
}

// convertToBoolSlice converts a slice to []bool.
func convertToBoolSlice(input []interface{}) ([]bool, bool) {
	result := make([]bool, 0, len(input))
	for _, v := range input {
		result = append(result, toBool(v))
	}
	return result, true
}

func toString(value interface{}) string {
	switch v := value.(type) {
	case string:
		return v
	case int, int8, int16, int32, int64:
		return fmt.Sprintf("%d", v)
	case uint, uint8, uint16, uint32, uint64:
		return fmt.Sprintf("%d", v)
	case float32, float64:
		return fmt.Sprintf("%f", v)
	case bool:
		return fmt.Sprintf("%t", v)
	default:
		return ""
	}
}

func toInt(value interface{}) int {
	switch v := value.(type) {
	case int:
		return v
	case int64:
		return int(v)
	case int32:
		return int(v)
	case uint:
		return int(v)
	case uint64:
		return int(v)
	case uint32:
		return int(v)
	case float64:
		if v == float64(int(v)) {
			return int(v)
		}
		return 0
	case float32:
		if v == float32(int(v)) {
			return int(v)
		}
		return 0
	case string:
		// Opcional: intentar parsear string a int
		i, err := strconv.Atoi(v)
		if err == nil {
			return i
		}
		return 0
	default:
		return 0
	}
}

func toFloat64(value interface{}) float64 {
	switch v := value.(type) {
	case float64:
		return v
	case float32:
		return float64(v)
	case int:
		return float64(v)
	case int64:
		return float64(v)
	case int32:
		return float64(v)
	case uint:
		return float64(v)
	case uint64:
		return float64(v)
	case uint32:
		return float64(v)
	case string:
		// Opcional: intentar parsear string a float64
		f, err := strconv.ParseFloat(v, 64)
		if err == nil {
			return f
		}
		return 0
	default:
		return 0
	}
}

func toBool(value interface{}) bool {
	switch v := value.(type) {
	case bool:
		return v
	case string:
		if strings.EqualFold(v, "true") {
			return true
		} else if strings.EqualFold(v, "false") {
			return false
		}
		return false
	default:
		return false
	}
}
