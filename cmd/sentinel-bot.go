package cmd

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"gopkg.in/telebot.v3"

	"github.com/mexxo-dvp/sentinel-bot/pkg/metrics"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// the command that launches the Telegram bot
var sentinelBotCmd = &cobra.Command{
	Use:     "sentinel-bot",
	Aliases: []string{"start", "bot"},
	Short:   "Launches Telegram bot",
	Long:    "Launches Telegram bot (telebot). Requires TELE_TOKEN environment variable.",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("🔧 Running sentinel-bot version: %s\n", appVersion)

		// --- Telegram token ---
		teleToken := os.Getenv("TELE_TOKEN")
		if teleToken == "" {
			log.Fatal().Msg("TELE_TOKEN is not set")
		}

		timeoutSec := 10
		if v := os.Getenv("TELE_TIMEOUT_SEC"); v != "" {
			if n, err := strconv.Atoi(v); err == nil && n > 0 {
				timeoutSec = n
			}
		}

		pref := telebot.Settings{
			Token:  teleToken,
			Poller: &telebot.LongPoller{Timeout: time.Duration(timeoutSec) * time.Second},
		}

		bot, err := telebot.NewBot(pref)
		if err != nil {
			log.Fatal().Err(err).Msg("failed to create bot")
		}

		bot.Handle(telebot.OnText, func(c telebot.Context) error {
			msg := c.Text()

			// Start a span for message processing
			ctxSpan, span := otel.Tracer("sentinel-bot").Start(context.Background(), "incoming_message")
			defer span.End()

			// attributes for traces
			user := safeUser(c)
			chatID := fmt.Sprintf("%d", c.Chat().ID)
			span.SetAttributes(
				attribute.String("messaging.system", "telegram"),
				attribute.String("messaging.user", user),
				attribute.String("messaging.chat_id", chatID),
			)

			// metrics: count handled messages
			metrics.IncMessage(ctxSpan)

			// trace_id for logs
			traceID := trace.SpanContextFromContext(ctxSpan).TraceID().String()

			log.Info().
				Str("event", "incoming_message").
				Str("from_user", user).
				Str("chat_id", chatID).
				Str("text", msg).
				Str("trace_id", traceID).
				Msg("Message received")

			return c.Send("You wrote: " + msg)
		})

		log.Info().Msg("✅ Bot is running. Waiting for messages...")
		bot.Start()
	},
}

func init() {
	rootCmd.AddCommand(sentinelBotCmd)
}

func getenvDefault(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func safeUser(c telebot.Context) string {
	if c.Sender() == nil {
		return "unknown"
	}
	u := c.Sender()
	if u.Username != "" {
		return u.Username
	}
	return fmt.Sprintf("%s %s (%d)", u.FirstName, u.LastName, u.ID)
}
