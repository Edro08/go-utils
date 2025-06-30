package logger

import (
	"fmt"
	"strings"
)

func formatKeys(keys []any) string {
	if len(keys) == 0 {
		return ""
	}

	var sb strings.Builder
	for i := 0; i < len(keys); i += 2 {
		// Procesar clave
		var key string
		if k, ok := keys[i].(string); ok && k != "" {
			key = k
		} else {
			key = fmt.Sprintf("<key:%v>", keys[i])
		}

		// Procesar valor
		val := "<missing>"
		if i+1 < len(keys) {
			if keys[i+1] != nil {
				val = fmt.Sprintf("%v", keys[i+1])
			} else {
				val = "<nil>"
			}
		}

		sb.WriteString(fmt.Sprintf(" | %s = %s", key, val))
	}

	return sb.String()
}
