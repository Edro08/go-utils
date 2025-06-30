package logger

import (
	"encoding/json"
	"fmt"
	"runtime"
	"time"
)

// Entry representa una entrada de log con metadatos como nivel, timestamp, ubicación, título y valores clave.
type Entry struct {
	Level     string    `json:"level"`     // Nivel del log (DEBUG, INFO, etc.)
	Timestamp time.Time `json:"timestamp"` // Momento en que se generó el log
	Location  string    `json:"location"`  // Archivo y línea de donde se generó el log
	Title     string    `json:"title"`     // Mensaje principal o título del log
	KeyVals   []any     `json:"keysVals"`  // Pares clave-valor opcionales para contexto adicional
}

// write construye y registra una entrada de log si el nivel es igual o superior al mínimo configurado.
// Determina la ubicación del llamado (archivo y línea), construye la entrada y la delega al formateador correspondiente.
//
// Parámetros:
//   - level: Nivel de log (DEBUG, INFO, WARN, ERROR, FATAL)
//   - title: Mensaje principal del log
//   - keysVals: Valores adicionales opcionales como pares clave-valor
func (l *Logger) write(level Level, title string, keysVals ...any) {
	if !(level >= l.opts.MinLevel) {
		return
	}

	_, file, line, ok := runtime.Caller(2)
	if !ok {
		file = "unknown"
	}

	e := Entry{
		Level:     levelLabels[level],
		Timestamp: time.Now(),
		Location:  fmt.Sprintf("%s:%d", file, line),
		Title:     title,
		KeyVals:   keysVals,
	}

	switch l.opts.Format {
	case FormatJSON:
		l.writeJSON(&e)
	case FormatText:
		l.writeText(&e)
	}
}

// writeJSON serializa una entrada de log al formato JSON y la imprime en consola.
// Si ocurre un error al serializar, se descarta silenciosamente.
//
// Parámetros:
//   - entry: Entrada de log a serializar
func (l *Logger) writeJSON(entry *Entry) {
	data, err := json.Marshal(entry)
	if err != nil {
		return
	}

	fmt.Println(string(data))
}

// writeText imprime una entrada de log en formato de texto legible con timestamp, ubicación y mensaje.
//
// Formato de salida: LEVEL [timestamp] (ubicación) título clave1=valor1 clave2=valor2 ...
//
// Parámetros:
//   - e: Entrada de log a formatear
func (l *Logger) writeText(e *Entry) {
	line := fmt.Sprintf("%s [%s] (%s) %s%s",
		e.Level,
		e.Timestamp.Format(time.RFC3339),
		e.Location,
		e.Title,
		formatKeys(e.KeyVals),
	)

	fmt.Println(line)
}
