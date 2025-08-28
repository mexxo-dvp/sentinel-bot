package cmd

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/spf13/cobra"
	"gopkg.in/telebot.v3"
)

// команда запуску Telegram-бота
var sentinelBotCmd = &cobra.Command{
	Use:     "sentinel-bot",        // ім’я CLI-команди ок
	Aliases: []string{"start", "bot"},
	Short:   "Запускає Telegram бота",
	Long:    "Запускає Telegram бота (telebot). Потрібна змінна середовища TELE_TOKEN.",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("🔧 Запуск sentinel-bot версії: %s\n", appVersion)

		teleToken := os.Getenv("TELE_TOKEN")
		if teleToken == "" {
			log.Fatal("❌ TELE_TOKEN не задано")
		}

		pref := telebot.Settings{
			Token:  teleToken,
			Poller: &telebot.LongPoller{Timeout: 10 * time.Second},
		}

		bot, err := telebot.NewBot(pref)
		if err != nil {
			log.Fatalf("❌ Не вдалося створити бота: %v", err)
		}

		bot.Handle(telebot.OnText, func(c telebot.Context) error {
			msg := c.Text()
			log.Printf("📩 %s", msg)
			return c.Send("Ти написав: " + msg)
		})

		fmt.Println("✅ Бот запущено. Очікування повідомлень…")
		bot.Start()
	},
}

func init() {
	rootCmd.AddCommand(sentinelBotCmd)
}
