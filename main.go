/*
Copyright © 2025 mexxo-dvp paranoidlookup@gmail.com
*/
package main

import (
	"context"
	"log"
	"os"

	"github.com/mexxo-dvp/sentinel-bot/cmd"
	"github.com/mexxo-dvp/sentinel-bot/pkg/logging"
	"github.com/mexxo-dvp/sentinel-bot/pkg/metrics"
	telemetry "github.com/mexxo-dvp/sentinel-bot/pkg/telemetry"
)

func main() {
	ctx := context.Background()

	// Service metadata from env (with defaults)
	serviceName := getenvDefault("OTEL_SERVICE_NAME", "sentinel-bot")
	serviceVersion := getenvDefault("APP_VERSION", "0.1.9")
	environment := getenvDefault("APP_ENV", os.Getenv("ENVIRONMENT"))

	// Initialize the logger: basic fields + level with LOG_LEVEL
	logging.Configure(serviceName, serviceVersion, environment)

	// Initialize OpenTelemetry (traces+metrics via OTLP) — SINGLE place
	prov, shutdown, err := telemetry.Init(ctx, serviceName, serviceVersion, environment)
	if err != nil {
		log.Printf("telemetry init failed: %v", err)
	} else {
		defer func() {
			if err := shutdown(ctx); err != nil {
				log.Printf("telemetry shutdown error: %v", err)
			}
		}()

		// Metrics init (IncMessage would be no-op without this)
		if err := metrics.Init(prov.Meter); err != nil {
			log.Printf("metrics init failed: %v", err)
		}
	}

	// Run Cobra
	cmd.Execute()
}

func getenvDefault(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}
