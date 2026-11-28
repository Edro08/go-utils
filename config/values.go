package config

import (
	"fmt"
	"math"
	"os"
	"reflect"
	"strconv"
	"strings"
	"time"
)

func (c *Config) separator() string {
	if c.opts.Separator == "" {
		return defaultSeparator
	}
	return c.opts.Separator
}

// cloneValue realiza una copia profunda de los valores internos. Previene que mapas, slices y []byte retornados mantengan referencias a c.data.
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
	case []byte:
		cp := make([]byte, len(val))
		copy(cp, val)
		return cp
	default:
		// Los tipos primitivos (string, int, bool, time.Duration, etc.) en Go se pasan por valor.
		return val
	}
}

// normalizeValue normaliza valores externos a la representación interna utilizada por Config.
// Los mapas se convierten a map[string]interface{} y los slices/arrays (excepto []byte) a []interface{} de forma recursiva.
func normalizeValue(v interface{}) (interface{}, error) {
	if v == nil {
		return nil, nil
	}

	switch typed := v.(type) {
	case []byte:
		cp := make([]byte, len(typed))
		copy(cp, typed)
		return cp, nil
	case time.Duration:
		return typed, nil
	}

	rv := reflect.ValueOf(v)

	switch rv.Kind() {
	case reflect.Map:
		result := make(map[string]interface{}, rv.Len())

		for _, key := range rv.MapKeys() {
			k := key
			if k.Kind() == reflect.Interface && !k.IsNil() {
				k = k.Elem()
			}
			if k.Kind() != reflect.String {
				return nil, fmt.Errorf("%w: got %s", ErrMapKeyType, k.Kind())
			}

			value, err := normalizeValue(
				rv.MapIndex(key).Interface(),
			)
			if err != nil {
				return nil, err
			}

			result[k.String()] = value
		}

		return result, nil

	case reflect.Slice, reflect.Array:
		result := make([]interface{}, rv.Len())

		for i := 0; i < rv.Len(); i++ {
			value, err := normalizeValue(rv.Index(i).Interface())
			if err != nil {
				return nil, err
			}

			result[i] = value
		}

		return result, nil

	default:
		return v, nil
	}
}

// envKeyName construye el nombre de variable de entorno correspondiente a una clave jerárquica.
func (c *Config) envKeyName(key string) string {
	sep := c.separator()
	normalized := strings.ReplaceAll(key, sep, "_")
	normalized = strings.ReplaceAll(normalized, ".", "_")
	normalized = strings.ReplaceAll(normalized, "-", "_")
	normalized = strings.ToUpper(normalized)

	prefix := strings.TrimSpace(c.opts.EnvPrefix)
	prefix = strings.TrimSuffix(prefix, "_")
	if prefix != "" {
		return strings.ToUpper(prefix) + "_" + normalized
	}
	return normalized
}

func (c *Config) lookupEnv(key string) (string, bool) {
	if !c.opts.AutomaticEnv || key == "" {
		return "", false
	}
	return os.LookupEnv(c.envKeyName(key))
}

func (c *Config) applyEnvOverridesToMap(prefix string, m map[string]interface{}) {
	sep := c.separator()
	for k, v := range m {
		fullKey := k
		if prefix != "" {
			fullKey = prefix + sep + k
		}
		if sub, ok := v.(map[string]interface{}); ok {
			c.applyEnvOverridesToMap(fullKey, sub)
			continue
		}
		if envVal, ok := c.lookupEnv(fullKey); ok {
			m[k] = coerceEnvValue(v, envVal)
		}
	}
}

func coerceEnvValue(existing interface{}, envVal string) interface{} {
	switch existing.(type) {
	case []interface{}:
		return splitEnvSlice(envVal)
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
		if i, ok := parseInt(envVal); ok {
			return i
		}
		return envVal
	case float32, float64:
		if f, ok := parseFloat64(envVal); ok {
			return f
		}
		return envVal
	case bool:
		if b, ok := parseBool(envVal); ok {
			return b
		}
		return envVal
	case time.Duration:
		if d, ok := parseDuration(envVal); ok {
			return d
		}
		return envVal
	default:
		return envVal
	}
}

func splitEnvSlice(raw string) []interface{} {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return []interface{}{}
	}
	parts := strings.Split(trimmed, ",")
	res := make([]interface{}, 0, len(parts))
	for _, p := range parts {
		res = append(res, strings.TrimSpace(p))
	}
	return res
}

// findValue navega el mapa interno (y variables de entorno si AutomaticEnv está activo).
// Debe ser llamado únicamente bajo un bloqueo (RLock o Lock).
func (c *Config) findValue(key string) (interface{}, bool) {
	if key == "" {
		return nil, false
	}

	if envVal, ok := c.lookupEnv(key); ok {
		return envVal, true
	}

	keys := strings.Split(key, c.separator())
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

// getValue busca el valor y lo clona antes de devolverlo. Debe llamarse bajo RLock/Lock.
func (c *Config) getValue(key string) (interface{}, bool) {
	val, ok := c.findValue(key)
	if !ok {
		return nil, false
	}

	return cloneValue(val), true
}

// getMap devuelve una copia segura del mapa interno.
func (c *Config) getMap(key string) (map[string]interface{}, bool) {
	v, ok := c.getValue(key)
	if !ok {
		return nil, false
	}

	m, ok := v.(map[string]interface{})
	return m, ok
}

// getSlice devuelve una copia segura del slice interno (o parsea lista separada por comas desde AutomaticEnv).
func (c *Config) getSlice(key string) ([]interface{}, bool) {
	if envVal, ok := c.lookupEnv(key); ok {
		return splitEnvSlice(envVal), true
	}

	v, ok := c.getValue(key)
	if !ok {
		return nil, false
	}

	s, ok := v.([]interface{})
	return s, ok
}

func (c *Config) set(key string, value interface{}) error {
	if key == "" {
		return ErrKeyEmpty
	}

	value, err := normalizeValue(value)
	if err != nil {
		return err
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	if c.data == nil {
		c.data = make(map[string]interface{})
	}

	keys := strings.Split(key, c.separator())
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

	cm[keys[len(keys)-1]] = value

	return nil
}

// convertToStringMap convierte un map[string]interface{} a map[string]string,
// aplicando la conversión a string para cada valor.
func convertToStringMap(input map[string]interface{}) (map[string]string, bool) {
	data := make(map[string]string, len(input))
	for k, v := range input {
		data[k] = toString(v)
	}
	return data, true
}

// convertToIntMap convierte un map[string]interface{} a map[string]int,
// aplicando la conversión a int para cada valor.
func convertToIntMap(input map[string]interface{}) (map[string]int, bool) {
	data := make(map[string]int, len(input))
	for k, v := range input {
		data[k] = toInt(v)
	}
	return data, true
}

// convertToFloatMap convierte un map[string]interface{} a map[string]float64,
// aplicando la conversión a float64 para cada valor.
func convertToFloatMap(input map[string]interface{}) (map[string]float64, bool) {
	data := make(map[string]float64, len(input))
	for k, v := range input {
		data[k] = toFloat64(v)
	}
	return data, true
}

// convertToBoolMap convierte un map[string]interface{} a map[string]bool,
// aplicando la conversión a bool para cada valor.
func convertToBoolMap(input map[string]interface{}) (map[string]bool, bool) {
	data := make(map[string]bool, len(input))
	for k, v := range input {
		data[k] = toBool(v)
	}
	return data, true
}

// convertToStringSlice convierte un slice []interface{} a []string,
// aplicando la conversión a string para cada elemento.
func convertToStringSlice(input []interface{}) ([]string, bool) {
	result := make([]string, 0, len(input))
	for _, v := range input {
		result = append(result, toString(v))
	}
	return result, true
}

// convertToIntSlice convierte un slice []interface{} a []int,
// aplicando la conversión a int para cada elemento.
func convertToIntSlice(input []interface{}) ([]int, bool) {
	result := make([]int, 0, len(input))
	for _, v := range input {
		result = append(result, toInt(v))
	}
	return result, true
}

// convertToFloatSlice convierte un slice []interface{} a []float64,
// aplicando la conversión a float64 para cada elemento.
func convertToFloatSlice(input []interface{}) ([]float64, bool) {
	result := make([]float64, 0, len(input))
	for _, v := range input {
		result = append(result, toFloat64(v))
	}
	return result, true
}

// convertToBoolSlice convierte un slice []interface{} a []bool,
// aplicando la conversión a bool para cada elemento.
func convertToBoolSlice(input []interface{}) ([]bool, bool) {
	result := make([]bool, 0, len(input))
	for _, v := range input {
		result = append(result, toBool(v))
	}
	return result, true
}

func parseString(value interface{}) (string, bool) {
	switch v := value.(type) {
	case string:
		return v, true
	case []byte:
		return string(v), true
	case time.Duration:
		return v.String(), true
	case int:
		return strconv.Itoa(v), true
	case int8:
		return strconv.FormatInt(int64(v), 10), true
	case int16:
		return strconv.FormatInt(int64(v), 10), true
	case int32:
		return strconv.FormatInt(int64(v), 10), true
	case int64:
		return strconv.FormatInt(v, 10), true
	case uint:
		return strconv.FormatUint(uint64(v), 10), true
	case uint8:
		return strconv.FormatUint(uint64(v), 10), true
	case uint16:
		return strconv.FormatUint(uint64(v), 10), true
	case uint32:
		return strconv.FormatUint(uint64(v), 10), true
	case uint64:
		return strconv.FormatUint(v, 10), true
	case float32:
		return strconv.FormatFloat(float64(v), 'f', -1, 32), true
	case float64:
		return strconv.FormatFloat(v, 'f', -1, 64), true
	case bool:
		return strconv.FormatBool(v), true
	default:
		return "", false
	}
}

// toString convierte cualquier valor básico a string,
// manejando tipos numéricos, booleanos, []byte, time.Duration y cadenas.
// Para tipos no reconocidos devuelve cadena vacía.
func toString(value interface{}) string {
	s, _ := parseString(value)
	return s
}

func parseInt(value interface{}) (int, bool) {
	switch v := value.(type) {
	case int:
		return v, true
	case int8:
		return int(v), true
	case int16:
		return int(v), true
	case int32:
		return int(v), true
	case int64:
		if strconv.IntSize == 32 && (v < math.MinInt32 || v > math.MaxInt32) {
			return 0, false
		}
		return int(v), true
	case uint:
		if strconv.IntSize == 32 && v > math.MaxInt32 {
			return 0, false
		}
		if strconv.IntSize == 64 && uint64(v) > math.MaxInt64 {
			return 0, false
		}
		return int(v), true
	case uint8:
		return int(v), true
	case uint16:
		return int(v), true
	case uint32:
		if strconv.IntSize == 32 && v > math.MaxInt32 {
			return 0, false
		}
		return int(v), true
	case uint64:
		if strconv.IntSize == 32 && v > math.MaxInt32 {
			return 0, false
		}
		if strconv.IntSize == 64 && v > math.MaxInt64 {
			return 0, false
		}
		return int(v), true
	case float64:
		if math.IsNaN(v) || math.IsInf(v, 0) || v != math.Trunc(v) {
			return 0, false
		}
		if strconv.IntSize == 32 {
			if v < math.MinInt32 || v > math.MaxInt32 {
				return 0, false
			}
		} else {
			if v < math.MinInt64 || v >= float64(math.MaxInt64) {
				return 0, false
			}
		}
		return int(v), true

	case float32:
		f := float64(v)
		if math.IsNaN(f) || math.IsInf(f, 0) || f != math.Trunc(f) {
			return 0, false
		}
		if strconv.IntSize == 32 {
			if f < math.MinInt32 || f > math.MaxInt32 {
				return 0, false
			}
		} else {
			if f < math.MinInt64 || f >= float64(math.MaxInt64) {
				return 0, false
			}
		}
		return int(v), true

	case string:
		i, err := strconv.Atoi(strings.TrimSpace(v))
		if err != nil {
			return 0, false
		}
		return i, true

	default:
		return 0, false
	}
}

// toInt convierte un valor básico a int, realizando conversiones
// seguras desde tipos numéricos y cadenas. Retorna 0 si no es convertible.
func toInt(value interface{}) int {
	i, _ := parseInt(value)
	return i
}

func parseFloat64(value interface{}) (float64, bool) {
	switch v := value.(type) {
	case float64:
		if math.IsNaN(v) || math.IsInf(v, 0) {
			return 0, false
		}
		return v, true
	case float32:
		f := float64(v)
		if math.IsNaN(f) || math.IsInf(f, 0) {
			return 0, false
		}
		return f, true
	case int:
		return float64(v), true
	case int8:
		return float64(v), true
	case int16:
		return float64(v), true
	case int32:
		return float64(v), true
	case int64:
		return float64(v), true
	case uint:
		return float64(v), true
	case uint8:
		return float64(v), true
	case uint16:
		return float64(v), true
	case uint32:
		return float64(v), true
	case uint64:
		return float64(v), true
	case string:
		f, err := strconv.ParseFloat(strings.TrimSpace(v), 64)
		if err == nil && !math.IsNaN(f) && !math.IsInf(f, 0) {
			return f, true
		}
		return 0, false
	default:
		return 0, false
	}
}

// toFloat64 convierte un valor básico a float64, incluyendo conversiones
// desde enteros, flotantes y cadenas.
// Retorna 0 si no es convertible.
func toFloat64(value interface{}) float64 {
	f, _ := parseFloat64(value)
	return f
}

func parseBool(value interface{}) (bool, bool) {
	switch v := value.(type) {
	case bool:
		return v, true
	case string:
		s := strings.TrimSpace(v)
		if strings.EqualFold(s, "true") {
			return true, true
		}
		if strings.EqualFold(s, "false") {
			return false, true
		}
		return false, false
	default:
		return false, false
	}
}

// toBool convierte un valor básico a bool, aceptando valores booleanos
// y cadenas "true"/"false" (sin distinguir mayúsculas).
// Retorna false si no es convertible.
func toBool(value interface{}) bool {
	b, _ := parseBool(value)
	return b
}

func parseDuration(value interface{}) (time.Duration, bool) {
	switch v := value.(type) {
	case time.Duration:
		return v, true
	case string:
		d, err := time.ParseDuration(strings.TrimSpace(v))
		if err != nil {
			return 0, false
		}
		return d, true
	default:
		return 0, false
	}
}
