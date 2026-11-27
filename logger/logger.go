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

// Slog retorna la instancia interna de *slog.Logger de la librería estándar.
func (l *Logger) Slog() *slog.Logger {
	return l.slog
}

// With retorna un nuevo ILogger derivado que incluye atributos contextuales fijos.
func (l *Logger) With(args ...any) ILogger {
	return &Logger{
		opts:   l.opts,
		slog:   l.slog.With(args...),
		closer: l.closer,
		exit:   l.exit,
	}
}

// WithGroup retorna un nuevo ILogger derivado que agrupa los atributos bajo un grupo.
func (l *Logger) WithGroup(name string) ILogger {
	return &Logger{
		opts:   l.opts,
		slog:   l.slog.WithGroup(name),
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

// writeRecord delega la creación del registro a log/slog inyectando atributos de contexto si aplica.
func (l *Logger) writeRecord(ctx context.Context, level Level, pc uintptr, title string, keysVals ...any) {
	if ctx == nil {
		ctx = context.Background()
	}

	slogLevel := level.ToSlog()
	if !l.slog.Enabled(ctx, slogLevel) {
		return
	}

	record := slog.NewRecord(time.Now(), slogLevel, title, pc)
	if l.opts.ContextExtractor != nil {
		if attrs := l.opts.ContextExtractor(ctx); len(attrs) > 0 {
			record.AddAttrs(attrs...)
		}
	}
	record.Add(keysVals...)

	_ = l.slog.Handler().Handle(ctx, record)
}
