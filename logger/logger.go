package logger

import (
	"context"
	"log/slog"
	"os"
	"runtime"
	"time"
)

// ILogger define el contrato de logging estructurado para inyección de dependencias.
// ------------------------------------------------------------------------------------------------
type ILogger interface {
	Debug(title string, keys ...any)
	Info(title string, keys ...any)
	Warn(title string, keys ...any)
	Error(title string, keys ...any)
	Fatal(title string, keys ...any)

	DebugCtx(ctx context.Context, title string, keys ...any)
	InfoCtx(ctx context.Context, title string, keys ...any)
	WarnCtx(ctx context.Context, title string, keys ...any)
	ErrorCtx(ctx context.Context, title string, keys ...any)
	FatalCtx(ctx context.Context, title string, keys ...any)

	Slog() *slog.Logger
	With(args ...any) ILogger
	WithGroup(name string) ILogger
	Close() error
}

// ------------------------------------------------------------------------------------------------
// Implementation Methods
// ------------------------------------------------------------------------------------------------

func (l *Logger) Debug(title string, keys ...any) {
	l.writeRecord(context.Background(), DEBUG, l.callerPC(), title, keys...)
}

func (l *Logger) Info(title string, keys ...any) {
	l.writeRecord(context.Background(), INFO, l.callerPC(), title, keys...)
}

func (l *Logger) Warn(title string, keys ...any) {
	l.writeRecord(context.Background(), WARN, l.callerPC(), title, keys...)
}

func (l *Logger) Error(title string, keys ...any) {
	l.writeRecord(context.Background(), ERROR, l.callerPC(), title, keys...)
}

func (l *Logger) Fatal(title string, keys ...any) {
	l.writeRecord(context.Background(), FATAL, l.callerPC(), title, keys...)
	_ = l.Close()
	l.exitProcess(1)
}

func (l *Logger) DebugCtx(ctx context.Context, title string, keys ...any) {
	l.writeRecord(ctx, DEBUG, l.callerPC(), title, keys...)
}

func (l *Logger) InfoCtx(ctx context.Context, title string, keys ...any) {
	l.writeRecord(ctx, INFO, l.callerPC(), title, keys...)
}

func (l *Logger) WarnCtx(ctx context.Context, title string, keys ...any) {
	l.writeRecord(ctx, WARN, l.callerPC(), title, keys...)
}

func (l *Logger) ErrorCtx(ctx context.Context, title string, keys ...any) {
	l.writeRecord(ctx, ERROR, l.callerPC(), title, keys...)
}

func (l *Logger) FatalCtx(ctx context.Context, title string, keys ...any) {
	l.writeRecord(ctx, FATAL, l.callerPC(), title, keys...)
	_ = l.Close()
	l.exitProcess(1)
}

func (l *Logger) exitProcess(code int) {
	if l.exit != nil {
		l.exit(code)
		return
	}
	os.Exit(code)
}

// Slog retorna la instancia interna de *slog.Logger de la librería estándar,
// preservando atributos y grupos derivados si existen.
func (l *Logger) Slog() *slog.Logger {
	s := l.slog
	if len(l.attrs) > 0 {
		anyAttrs := make([]any, len(l.attrs))
		for i, a := range l.attrs {
			anyAttrs[i] = a
		}
		s = s.With(anyAttrs...)
	}
	for _, g := range l.groups {
		s = s.WithGroup(g)
	}
	return s
}

// With retorna un nuevo ILogger derivado cuyos atributos fijos se agregan al final de cada registro.
func (l *Logger) With(args ...any) ILogger {
	added := wrapGroups(l.groups, argsToAttrs(args...))
	if len(added) == 0 {
		return l
	}

	nextAttrs := make([]slog.Attr, 0, len(l.attrs)+len(added))
	nextAttrs = append(nextAttrs, l.attrs...)
	nextAttrs = append(nextAttrs, added...)

	return &Logger{
		opts:   l.opts,
		slog:   l.slog,
		attrs:  nextAttrs,
		groups: l.groups,
		closer: l.closer,
		exit:   l.exit,
	}
}

// WithGroup retorna un nuevo ILogger derivado que agrupa los atributos subsecuentes bajo un grupo.
func (l *Logger) WithGroup(name string) ILogger {
	if name == "" {
		return l
	}

	nextGroups := make([]string, 0, len(l.groups)+1)
	nextGroups = append(nextGroups, l.groups...)
	nextGroups = append(nextGroups, name)

	return &Logger{
		opts:   l.opts,
		slog:   l.slog,
		attrs:  l.attrs,
		groups: nextGroups,
		closer: l.closer,
		exit:   l.exit,
	}
}

// ToSlog convierte el Level interno al Level de log/slog.
func (l Level) ToSlog() slog.Level {
	switch l {
	case DEBUG:
		return slog.LevelDebug
	case INFO:
		return slog.LevelInfo
	case WARN:
		return slog.LevelWarn
	case ERROR:
		return slog.LevelError
	case FATAL:
		return fatalSlogLevel
	default:
		return slog.LevelInfo
	}
}

// callerPC obtiene el program counter del punto de llamada público únicamente si AddSource está habilitado.
func (l *Logger) callerPC() uintptr {
	if !l.opts.AddSource {
		return 0
	}
	var pcs [1]uintptr
	// [0] runtime.Callers, [1] l.callerPC, [2] método público (Info/Debug/etc.), [3] código del llamador
	runtime.Callers(3, pcs[:])
	return pcs[0]
}

// write mantiene compatibilidad interna para escritura sin contexto explícito.
func (l *Logger) write(level Level, title string, keysVals ...any) {
	var pc uintptr
	if l.opts.AddSource {
		var pcs [1]uintptr
		runtime.Callers(2, pcs[:])
		pc = pcs[0]
	}
	l.writeRecord(context.Background(), level, pc, title, keysVals...)
}

// writeRecord delega la creación del registro a log/slog colocando primero los atributos del evento
// y al final los atributos de contexto y del logger derivado (With).
func (l *Logger) writeRecord(ctx context.Context, level Level, pc uintptr, title string, keysVals ...any) {
	if ctx == nil {
		ctx = context.Background()
	}

	slogLevel := level.ToSlog()
	if !l.slog.Enabled(ctx, slogLevel) {
		return
	}

	record := slog.NewRecord(time.Now(), slogLevel, title, pc)
	if len(keysVals) > 0 {
		if len(l.groups) == 0 {
			record.Add(keysVals...)
		} else {
			record.AddAttrs(wrapGroups(l.groups, argsToAttrs(keysVals...))...)
		}
	}
	if l.opts.ContextExtractor != nil {
		if attrs := l.opts.ContextExtractor(ctx); len(attrs) > 0 {
			record.AddAttrs(attrs...)
		}
	}
	if len(l.attrs) > 0 {
		record.AddAttrs(l.attrs...)
	}

	_ = l.slog.Handler().Handle(ctx, record)
}

func argsToAttrs(args ...any) []slog.Attr {
	if len(args) == 0 {
		return nil
	}
	var r slog.Record
	r.Add(args...)
	attrs := make([]slog.Attr, 0, r.NumAttrs())
	r.Attrs(func(a slog.Attr) bool {
		attrs = append(attrs, a)
		return true
	})
	return attrs
}

func wrapGroups(groups []string, attrs []slog.Attr) []slog.Attr {
	if len(groups) == 0 || len(attrs) == 0 {
		return attrs
	}
	grouped := attrs
	for i := len(groups) - 1; i >= 0; i-- {
		anyAttrs := make([]any, len(grouped))
		for j, a := range grouped {
			anyAttrs[j] = a
		}
		grouped = []slog.Attr{slog.Group(groups[i], anyAttrs...)}
	}
	return grouped
}
