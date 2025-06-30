package logger

import (
	"fmt"
	"strings"
)

// formatKeys convierte una lista de elementos (esperados como pares clave-valor) en una cadena formateada.
// Cada par se muestra en el formato: " | clave = valor". Si falta el valor o es nil, se reemplaza con etiquetas "<missing>" o "<nil>".
//
// Parámetros:
//   - keys: Slice de elementos alternando clave (string) y valor (any)
//
// Retorna:
//   - Una cadena con todos los pares clave-valor concatenados en formato legible.
//     Si el slice está vacío, se retorna una cadena vacía.
//
// Ejemplo de salida:
//
//	" | usuario = juan | activo = true | edad = 30"
func formatKeys(keys []any) string {
	if len(keys) == 0 {
		return ""
	}

	var sb strings.Builder
	for i := 0; i < len(keys); i += 2 {
		var key string
		if k, ok := keys[i].(string); ok && k != "" {
			key = k
		} else {
			key = fmt.Sprintf("<key:%v>", keys[i])
		}

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
