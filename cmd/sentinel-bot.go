package cmd

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/hirokisan/zerodriver"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/mexxo-dvp/sentinel-bot/pkg/logging"
	"github.com/mexxo-dvp/sentinel-bot/pkg/telemetry"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"gopkg.in/telebot.v3"
)

var (
	httpAddr string
)

func init() {
	rootCmd.AddCommand(sentinelBotCmd)
	addBotFlags(sentinelBotCmd.Flags())
}

func addBotFlags(fs *pflag.FlagSet) {
	fs.StringVar(&httpAddr, "http-addr", ":8080", "HTTP address for health")
}

// command to launch Telegram bot
var sentinelBotCmd = &cobra.Command{
	Use:     "sentinel-bot",
	Aliases: []string{"start", "bot"},
	Short:   "Запускає Telegram бота",
	Long:    "Запускає Telegram бота (telebot). Потрібна змінна середовища TELE_TOKEN.",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Printf("🔧 Запуск sentinel-bot версії: %s\n", appVersion)

		// --- init logging
		logging.Init()

		// --- init telemetry
		ctx := context.Background()
		env := getenv("APP_ENV", "dev")
		otelProviders, shutdown, err := telemetry.Init(ctx, "sentinel-bot", appVersion, env)
		logging.FatalIf(err, "init otel")

		defer func() {
			_ = shutdown(context.Background())
		}()

		// metric: inbound text messages counter (with instances tied to an active spawn)
		counter, err := otelProviders.Meter.Int64Counter("sentinelbot_messages_total",
			metric.WithDescription("Total number of inbound text messages"),
			metric.WithUnit("{message}"),
		)
		logging.FatalIf(err, "create metric")

		teleToken := os.Getenv("TELE_TOKEN")
		if teleToken == "" {
			return fmt.Errorf("TELE_TOKEN не задано")
		}

		pref := telebot.Settings{
			Token:  teleToken,
			Poller: &telebot.LongPoller{Timeout: 10 * time.Second},
		}

		bot, err := telebot.NewBot(pref)
		if err != nil {
			return fmt.Errorf("bot creation failed: %w", err)
		}

		// --- health HTTP
		go func() {
			mux := http.NewServeMux()
			mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
				_, _ = w.Write([]byte("ok"))
			})
			_ = http.ListenAndServe(httpAddr, mux)
		}()

		tr := otel.Tracer("sentinel-bot/telegram")

		bot.Handle(telebot.OnText, func(c telebot.Context) error {
			// create a span for an event from Telegram (SpanKindConsumer)
			ctx, span := tr.Start(c.Context(), "telegram.on_text",
				trace.WithSpanKind(trace.SpanKindConsumer))
			defer span.End()

			msg := c.Text()
			user := c.Sender()

			// log in JSON with trace_id/span_id
			logging.InfoCtx(ctx, "inbound message",
				zerodriver.String("payload", msg),
				zerodriver.String("user", fmt.Sprintf("%s(%d)", user.Username, user.ID)),
				zerodriver.String("severity", "INFO"),
			)

			// increment the metric with the instance (trace exemplar)
			counter.Add(ctx, 1,
				attribute.String("message_kind", "text"),
			)

			return c.Send("You wrote: " + msg)
		})

		fmt.Println("✅ Bot is running. Waiting for messages...")
		bot.Start()
		return nil
	},
}

func getenv(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}
