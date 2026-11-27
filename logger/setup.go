/*
Package logger provee un logger estructurado basado en la biblioteca estándar log/slog para aplicaciones escritas en Go.

Permite registrar mensajes en diferentes niveles, utilizar formato JSON o texto
y configurar la salida hacia consola (stdout), archivo o ambos.

Uso básico:

```

	log, err := logger.New(logger.Options{
		MinLevel: logger.INFO,
		Format:   logger.FormatJSON,
	})

	if err != nil {
		panic(err)
	}
	defer log.Close()

log.Info("server started", "port", 8080)
log.Warn("high memory usage", "usage", "85%")
log.Error("failed to process request", "error", err)
```

Para asociar atributos contextuales o grupos a un logger:

```
requestLogger := log.With(

	"request_id", requestID,
	"user_id", userID,

)

requestLogger.Info("request processed")
```

Comportamiento:
  - Por defecto, Options{} inicializa MinLevel en INFO, Format en FormatJSON, Output en Stdout y AddSource en false.
  - Los registros por debajo de MinLevel son ignorados.
  - AddSource está desactivado por defecto para evitar overhead y filtración de rutas absolutas; si se activa, recorta la ruta a formato corto (paquete/archivo.go).
  - Permite utilizar los niveles DEBUG, INFO, WARN, ERROR y FATAL.
  - FATAL registra el mensaje con etiqueta "FATAL", cierra el archivo de salida si existe y termina el proceso con código 1.
  - Permite registrar atributos estructurados mediante pares clave-valor y extraer atributos desde context.Context con ContextExtractor.
  - Permite asociar atributos permanentes mediante With y agruparlos mediante WithGroup retornando ILogger.
  - Cuando se habilita el registro en archivo y no se especifica una ruta con extensión .log,
    se genera un archivo con timestamp dentro del directorio indicado (o ./log por defecto).
*/
package logger

import (
	"context"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// Options define la configuración para inicializar un Logger.
type Options struct {
	MinLevel         Level
	Format           Format
	Output           Output
	EnableFile       bool
	FilePath         string
	AddSource        bool
	Writer           io.Writer
	ContextExtractor func(ctx context.Context) []slog.Attr
}

// Format define el formato de salida de los logs.
type Format string

const (
	FormatJSON Format = "JSON"
	FormatText Format = "TEXT"
)

// ParseFormat convierte una cadena ("JSON", "TEXT") en Format sin distinguir mayúsculas.
func ParseFormat(s string) (Format, bool) {
	switch Format(strings.ToUpper(strings.TrimSpace(s))) {
	case FormatJSON:
		return FormatJSON, true
	case FormatText:
		return FormatText, true
	default:
		return FormatJSON, false
	}
}

// Output define el destino de escritura de los registros.
type Output string

const (
	Stdout Output = "STDOUT"
	File   Output = "FILE"
	Both   Output = "BOTH"
)

// Level representa los niveles de log soportados.
// Inicia en 1 para que el valor cero de Options{} adopte INFO como nivel por defecto.
type Level int

const (
	DEBUG Level = iota + 1
	INFO
	WARN
	ERROR
	FATAL
)

const fatalSlogLevel = slog.LevelError + 4

// String devuelve la representación textual del nivel.
func (l Level) String() string {
	switch l {
	case DEBUG:
		return "DEBUG"
	case INFO:
		return "INFO"
	case WARN:
		return "WARN"
	case ERROR:
		return "ERROR"
	case FATAL:
		return "FATAL"
	default:
		return "INFO"
	}
}

// ParseLevel convierte una cadena en Level sin distinguir mayúsculas.
func ParseLevel(s string) (Level, bool) {
	switch strings.ToUpper(strings.TrimSpace(s)) {
	case "DEBUG":
		return DEBUG, true
	case "INFO":
		return INFO, true
	case "WARN", "WARNING":
		return WARN, true
	case "ERROR":
		return ERROR, true
	case "FATAL":
		return FATAL, true
	default:
		return INFO, false
	}
}

// onceCloser garantiza que el cierre de recursos compartidos (ej. entre loggers derivados con With) sea idempotente y seguro en concurrencia.
type onceCloser struct {
	once   sync.Once
	closer io.Closer
	err    error
}

func (o *onceCloser) Close() error {
	if o == nil || o.closer == nil {
		return nil
	}
	o.once.Do(func() {
		o.err = o.closer.Close()
	})
	return o.err
}

// Logger es el wrapper estructurado sobre log/slog.
type Logger struct {
	opts   Options
	slog   *slog.Logger
	closer io.Closer
	exit   func(int)
}

// New crea e inicializa una nueva instancia de Logger basado en log/slog.
func New(opts Options) (*Logger, error) {
	validateOptions(&opts)

	w, closer, err := setupWriter(opts)
	if err != nil {
		return nil, err
	}

	if closer != nil {
		closer = &onceCloser{closer: closer}
	}

	handler := buildHandler(w, opts)

	return &Logger{
		opts:   opts,
		slog:   slog.New(handler),
		closer: closer,
	}, nil
}

// validateOptions ajusta valores por defecto para opciones no inicializadas o fuera de rango.
func validateOptions(opts *Options) {
	if opts.Format != FormatJSON && opts.Format != FormatText {
		opts.Format = FormatJSON
	}
	if opts.MinLevel < DEBUG || opts.MinLevel > FATAL {
		opts.MinLevel = INFO
	}

	switch opts.Output {
	case Stdout, File, Both:
		// Respetar Output explícito; si EnableFile fue activado junto a Stdout implícito, se mantiene Output explícito.
	default:
		if opts.EnableFile {
			opts.Output = Both
		} else {
			opts.Output = Stdout
		}
	}

	opts.EnableFile = opts.Output == File || opts.Output == Both
}

// setupWriter configura la salida del logger según Options.Output, Options.EnableFile y Options.Writer.
func setupWriter(opts Options) (io.Writer, io.Closer, error) {
	console := io.Writer(os.Stdout)
	if opts.Writer != nil {
		console = opts.Writer
	}

	useFile := opts.Output == File || opts.Output == Both || opts.EnableFile
	if !useFile {
		return console, nil, nil
	}

	dir := opts.FilePath
	if dir == "" {
		dir = "./log"
	}

	var path string
	if strings.HasSuffix(strings.ToLower(dir), ".log") {
		path = dir
		dir = filepath.Dir(path)
	} else {
		timestamp := time.Now().Format("20060102T150405.000")
		path = filepath.Join(dir, timestamp+".log")
	}

	if dir != "" && dir != "." {
		if err := os.MkdirAll(dir, 0750); err != nil {
			return nil, nil, err
		}
	}

	file, err := os.OpenFile(
		path,
		os.O_CREATE|os.O_WRONLY|os.O_APPEND,
		0600,
	)
	if err != nil {
		return nil, nil, err
	}

	if opts.Output == File {
		return file, file, nil
	}

	return io.MultiWriter(console, file), file, nil
}

// buildHandler genera el slog.Handler correspondiente (JSON o Text).
func buildHandler(w io.Writer, opts Options) slog.Handler {
	handlerOpts := &slog.HandlerOptions{
		Level:     opts.MinLevel.ToSlog(),
		AddSource: opts.AddSource,
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			if len(groups) == 0 {
				if a.Key == slog.LevelKey {
					if lvl, ok := a.Value.Any().(slog.Level); ok && lvl >= fatalSlogLevel {
						return slog.String(slog.LevelKey, "FATAL")
					}
				}
				if opts.AddSource && a.Key == slog.SourceKey {
					if src, ok := a.Value.Any().(*slog.Source); ok && src != nil {
						cp := *src
						cp.File = shortSourceFile(src.File)
						return slog.Any(slog.SourceKey, &cp)
					}
				}
			}
			return a
		},
	}

	if opts.Format == FormatJSON {
		return slog.NewJSONHandler(w, handlerOpts)
	}
	return slog.NewTextHandler(w, handlerOpts)
}

// shortSourceFile recorta rutas absolutas a "directorio/archivo.go" (o "archivo.go") usando separadores '/'.
func shortSourceFile(file string) string {
	if file == "" {
		return ""
	}
	clean := filepath.ToSlash(file)
	idx := strings.LastIndexByte(clean, '/')
	if idx == -1 {
		return clean
	}
	prevIdx := strings.LastIndexByte(clean[:idx], '/')
	if prevIdx == -1 {
		part := clean[:idx]
		// Evitar prefijos de unidad tipo "C:" o cadenas vacías por "/" inicial.
		if part == "" || strings.HasSuffix(part, ":") {
			return clean[idx+1:]
		}
		return clean
	}
	return clean[prevIdx+1:]
}

func (l *Logger) Close() error {
	if l.closer == nil {
		return nil
	}

	return l.closer.Close()
}
