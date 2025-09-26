// CHANGED/NEW: logging package with Configure() function and saved helpers
package logging

import (
	"context"
	"os"
	"strings"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"go.opentelemetry.io/otel/trace"
)

var base zerolog.Logger

func init() {
	// ISO time in UTC — convenient for Loki
	zerolog.TimeFieldFormat = time.RFC3339Nano
	zerolog.TimestampFunc = func() time.Time { return time.Now().UTC() }
	base = zerolog.New(os.Stdout).With().Timestamp().Logger()
}

// Configure — NEW: sets up a global logger with basic fields and LOG_LEVEL
func Configure(service, version, environment string) {
	level := zerolog.InfoLevel
	if v := os.Getenv("LOG_LEVEL"); v != "" {
		if l, err := zerolog.ParseLevel(strings.ToLower(v)); err == nil {
			level = l
		}
	}
	zerolog.SetGlobalLevel(level)

	// basic fields in each record
	base = zerolog.New(os.Stdout).
		With().
		Timestamp().
		Str("service", service).
		Str("version", version).
		Str("env", environment).
		Logger()

	// Important: we are replacing the global logger of the log package.
	log.Logger = base
}

// --- Field helpers (compatible with previous code)
type Field func(*zerolog.Event) *zerolog.Event

func Str(k, v string) Field       { return func(e *zerolog.Event) *zerolog.Event { return e.Str(k, v) } }
func Int(k string, v int) Field   { return func(e *zerolog.Event) *zerolog.Event { return e.Int(k, v) } }
func Bool(k string, v bool) Field { return func(e *zerolog.Event) *zerolog.Event { return e.Bool(k, v) } }
func Err(err error) Field         { return func(e *zerolog.Event) *zerolog.Event { return e.Err(err) } }
func Any(k string, v any) Field   { return func(e *zerolog.Event) *zerolog.Event { return e.Any(k, v) } }

func withFields(e *zerolog.Event, fields []Field) *zerolog.Event {
	for _, f := range fields {
		e = f(e)
	}
	return e
}

func addTraceFromCtx(ctx context.Context, e *zerolog.Event) *zerolog.Event {
	sc := trace.SpanContextFromContext(ctx)
	if sc.IsValid() {
		e = e.
			Str("trace_id", sc.TraceID().String()).
			Str("span_id", sc.SpanID().String()).
			Bool("trace_sampled", sc.IsSampled())
	}
	return e
}

// --- Wrappers for the usual style
func Info(msg string, fields ...Field)  { withFields(base.Info(), fields).Msg(msg) }
func Error(msg string, fields ...Field) { withFields(base.Error(), fields).Msg(msg) }
func Debug(msg string, fields ...Field) { withFields(base.Debug(), fields).Msg(msg) }

func InfoCtx(ctx context.Context, msg string, fields ...Field) {
	e := addTraceFromCtx(ctx, base.Info())
	withFields(e, fields).Msg(msg)
}
func ErrorCtx(ctx context.Context, msg string, fields ...Field) {
	e := addTraceFromCtx(ctx, base.Error())
	withFields(e, fields).Msg(msg)
}
func DebugCtx(ctx context.Context, msg string, fields ...Field) {
	e := addTraceFromCtx(ctx, base.Debug())
	withFields(e, fields).Msg(msg)
}
