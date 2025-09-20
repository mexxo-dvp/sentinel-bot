package logging

import (
	"context"
	"os"
	"time"

	"github.com/rs/zerolog"
	"go.opentelemetry.io/otel/trace"
)

var base zerolog.Logger

func init() {
    // ISO time in UTC, convenient for Loki
	zerolog.TimeFieldFormat = time.RFC3339Nano
	zerolog.TimestampFunc = func() time.Time { return time.Now().UTC() }
	base = zerolog.New(os.Stdout).With().Timestamp().Logger()
}

// Field — functional option for adding fields (compatible with the usual style)
type Field func(*zerolog.Event) *zerolog.Event

func Str(k, v string) Field      { return func(e *zerolog.Event) *zerolog.Event { return e.Str(k, v) } }
func Int(k string, v int) Field  { return func(e *zerolog.Event) *zerolog.Event { return e.Int(k, v) } }
func Bool(k string, v bool) Field { return func(e *zerolog.Event) *zerolog.Event { return e.Bool(k, v) } }
func Err(err error) Field        { return func(e *zerolog.Event) *zerolog.Event { return e.Err(err) } }
func Any(k string, v any) Field  { return func(e *zerolog.Event) *zerolog.Event { return e.Any(k, v) } }

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

func Info(msg string, fields ...Field) {
	withFields(base.Info(), fields).Msg(msg)
}

func InfoCtx(ctx context.Context, msg string, fields ...Field) {
	e := base.Info()
	e = addTraceFromCtx(ctx, e)
	withFields(e, fields).Msg(msg)
}

func Error(msg string, fields ...Field) {
	withFields(base.Error(), fields).Msg(msg)
}

func ErrorCtx(ctx context.Context, msg string, fields ...Field) {
	e := base.Error()
	e = addTraceFromCtx(ctx, e)
	withFields(e, fields).Msg(msg)
}

func Debug(msg string, fields ...Field) {
	withFields(base.Debug(), fields).Msg(msg)
}

func DebugCtx(ctx context.Context, msg string, fields ...Field) {
	e := base.Debug()
	e = addTraceFromCtx(ctx, e)
	withFields(e, fields).Msg(msg)
}
