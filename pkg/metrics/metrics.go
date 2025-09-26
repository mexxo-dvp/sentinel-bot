package metrics

import (
	"context"

	"go.opentelemetry.io/otel/metric"
)

var msgCounter metric.Int64Counter // nil before initialization — Add() is not called

// Init — Call once after telemetry.Init(...).
// Without calling Init, increments will be no-op (safe behavior).
func Init(m metric.Meter) error {
	c, err := m.Int64Counter(
		"sentinel_messages_total",
		metric.WithDescription("Total number of handled Telegram messages"),
		metric.WithUnit("{message}"),
	)
	if err != nil {
		return err
	}
	msgCounter = c
	return nil
}

// IncMessage — call in the handler of each processed message.
// If Init has not been called yet, it is a no-op (does nothing).
func IncMessage(ctx context.Context) {
	if msgCounter != nil {
		msgCounter.Add(ctx, 1)
	}
}
