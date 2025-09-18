package logging

import (
	"context"
	"os"
	"time"

	"github.com/hirokisan/zerodriver"
	"go.opentelemetry.io/otel/trace"
)

var log *zerodriver.Logger

func Init() {
	l := zerodriver.NewProductionLogger()
	l = l.With(
		zerodriver.TimestampFunc(func() time.Time { return time.Now().UTC() }),
	)
	log = l
}

func WithTrace(ctx context.Context) *zerodriver.Event {
	ev := log.With()
	span := trace.SpanFromContext(ctx)
	sc := span.SpanContext()
	if sc.IsValid() {
		ev = ev.Str("trace_id", sc.TraceID().String()).Str("span_id", sc.SpanID().String())
	}
	return ev.Logger().Info()
}

func Info(msg string, fields ...zerodriver.Field) {
	log.Info().Fields(fields).Msg(msg)
}

func InfoCtx(ctx context.Context, msg string, fields ...zerodriver.Field) {
	ev := log.Info()
	span := trace.SpanFromContext(ctx)
	sc := span.SpanContext()
	if sc.IsValid() {
		ev = ev.Str("trace_id", sc.TraceID().String()).Str("span_id", sc.SpanID().String())
	}
	ev.Fields(fields).Msg(msg)
}

func ErrorCtx(ctx context.Context, msg string, fields ...zerodriver.Field) {
	ev := log.Error()
	span := trace.SpanFromContext(ctx)
	sc := span.SpanContext()
	if sc.IsValid() {
		ev = ev.Str("trace_id", sc.TraceID().String()).Str("span_id", sc.SpanID().String())
	}
	ev.Fields(fields).Msg(msg)
}

func FatalIf(err error, msg string) {
	if err != nil {
		log.Fatal().Err(err).Msg(msg)
	}
}

func Stdout() *os.File { return os.Stdout }
