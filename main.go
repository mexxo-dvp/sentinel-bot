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
	logging.Configure(serviceName, serviceVersion, environment) // NEW

	// Initializing OpenTelemetry (traces+metrics via OTLP)
	prov, shutdown, err := telemetry.Init(ctx, serviceName, serviceVersion, environment)
	if err != nil {
		log.Printf("telemetry init failed: %v", err)
	} else {
		defer func() {
			if err := shutdown(ctx); err != nil {
				log.Printf("telemetry shutdown error: %v", err)
			}
		}()

		// Initialization of metrics; without this, IncMessage will be no-op
		if err := metrics.Init(prov.Meter); err != nil {
			log.Printf("metrics init failed: %v", err)
		}
	}

	
	cmd.Execute()
}

func getenvDefault(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}
