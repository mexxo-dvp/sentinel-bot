package cmd

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/spf13/cobra"
	"gopkg.in/telebot.v3"
)

// sentinel-botCmd представляє команду запуску Telegram-бота
var kbotCmd = &cobra.Command{
	Use:     "sentinel-bot",
	Aliases: []string{"start"},
	Short:   "Запускає Telegram бота",
	Long: `Ця команда запускає Telegram бота з використанням бібліотеки telebot.
Потрібно, щоб була встановлена змінна середовища TELE_TOKEN.`,
	Run: func(cmd *cobra.Command, args []string) {
		// Виводимо версію
		fmt.Printf("🔧 Запуск sentinel-bot версії: %s\n", appVersion)

		// Отримуємо токен з середовища
		teleToken := os.Getenv("TELE_TOKEN")
		if teleToken == "" {
			log.Fatal("❌ Змінна середовища TELE_TOKEN не задана. Будь ласка, встановіть її перед запуском.")
		}

		// Налаштування бота
		pref := telebot.Settings{
			Token:  teleToken,
			Poller: &telebot.LongPoller{Timeout: 10 * time.Second},
		}

		// Ініціалізація бота
		bot, err := telebot.NewBot(pref)
		if err != nil {
			log.Fatalf("❌ Не вдалося створити бота: %v", err)
		}

		// Хендлер на вхідні текстові повідомлення
		bot.Handle(telebot.OnText, func(c telebot.Context) error {
			payload := c.Text()
			log.Printf("📩 Отримано повідомлення: %s", payload)
			return c.Send("Ти написав: " + payload)
		})

		// Запускаємо бота
		fmt.Println("✅ Бот запущено. Очікування повідомлень...")
		bot.Start()
	},
}

func init() {
	rootCmd.AddCommand(sentinel-botCmd)
}
