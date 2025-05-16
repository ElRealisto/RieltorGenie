// package bot

// import (
// 	"fmt"
// 	"log"
// 	"path/filepath"

// 	"github.com/ElRealisto/RieltorGenie/internal/objects"
// 	"github.com/ElRealisto/RieltorGenie/internal/users"
// 	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
// )

// // HandleCommand обробляє відомі команди
// func HandleCommand(bot *tgbotapi.BotAPI, update tgbotapi.Update) bool {
// 	if update.Message == nil {
// 		return false
// 	}

// 	switch update.Message.Text {
// 	case "/парсити":
// 		handleParseCommand(bot, update.Message)
// 		return true
// 	default:
// 		return false
// 	}
// }

// func handleParseCommand(bot *tgbotapi.BotAPI, message *tgbotapi.Message) {
// 	telegramID := message.From.ID

// 	user, found := users.FindByTelegramID(telegramID)
// 	if !found {
// 		msg := tgbotapi.NewMessage(message.Chat.ID, "🚫 Вас не знайдено у базі рієлторів.")
// 		bot.Send(msg)
// 		return
// 	}

// 	if user.Role != users.RealtorRole {
// 		msg := tgbotapi.NewMessage(message.Chat.ID, "🚫 Ця команда доступна лише рієлторам.")
// 		bot.Send(msg)
// 		return
// 	}

// 	msg := tgbotapi.NewMessage(message.Chat.ID, "🔍 Розпочато парсинг вашого профілю...")
// 	bot.Send(msg)

// 	properties, err := objects.ParseRealtorProfile(user.ProfileURL)
// 	if err != nil {
// 		log.Printf("❌ Помилка парсингу профілю %s: %v", user.ProfileURL, err)
// 		msg := tgbotapi.NewMessage(message.Chat.ID, "❌ Помилка при парсингу профілю.")
// 		bot.Send(msg)
// 		return
// 	}

// 	outputPath := filepath.Join("internal", "objects", "parsed_properties.json")
// 	err = objects.SavePropertiesToFile(properties, outputPath)
// 	if err != nil {
// 		log.Printf("❌ Помилка збереження JSON: %v", err)
// 		msg := tgbotapi.NewMessage(message.Chat.ID, "❌ Помилка при збереженні даних.")
// 		bot.Send(msg)
// 		return
// 	}

// 	log.Printf("✅ Успішно збережено %d об'єктів у %s", len(properties), outputPath)
// 	msg = tgbotapi.NewMessage(message.Chat.ID, fmt.Sprintf("✅ Збережено %d об'єктів!", len(properties)))
// 	bot.Send(msg)
// }

package bot

import (
	"fmt"
	"log"
	"path/filepath"

	"github.com/ElRealisto/RieltorGenie/internal/objects"
	"github.com/ElRealisto/RieltorGenie/internal/users"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// HandleUpdate обробляє вхідні повідомлення Telegram
func HandleUpdate(bot *tgbotapi.BotAPI, update tgbotapi.Update) {
	if update.Message == nil {
		return
	}

	switch update.Message.Text {
	case "/парсити":
		handleParseCommand(bot, update.Message)
	default:
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "🤖 Невідома команда. Спробуйте /парсити")
		bot.Send(msg)
	}
}

// handleParseCommand запускає парсинг для рієлтора
func handleParseCommand(bot *tgbotapi.BotAPI, message *tgbotapi.Message) {
	telegramID := message.From.ID

	user := users.FindByTelegramID(telegramID)
	if user == nil {
		msg := tgbotapi.NewMessage(message.Chat.ID, "🚫 Вас не знайдено у базі рієлторів.")
		bot.Send(msg)
		return
	}

	if user.Role != users.RealtorRole {
		msg := tgbotapi.NewMessage(message.Chat.ID, "🚫 Ця команда доступна лише рієлторам.")
		bot.Send(msg)
		return
	}

	msg := tgbotapi.NewMessage(message.Chat.ID, "🔍 Розпочато парсинг вашого профілю...")
	bot.Send(msg)

	properties, err := objects.ParseRealtorProfile(user.ProfileURL)
	if err != nil {
		log.Printf("❌ Помилка парсингу профілю %s: %v", user.ProfileURL, err)
		msg := tgbotapi.NewMessage(message.Chat.ID, "❌ Помилка при парсингу профілю.")
		bot.Send(msg)
		return
	}

	outputPath := filepath.Join("internal", "objects", "parsed_properties.json")
	err = objects.SavePropertiesToFile(properties, outputPath)
	if err != nil {
		log.Printf("❌ Помилка збереження JSON: %v", err)
		msg := tgbotapi.NewMessage(message.Chat.ID, "❌ Помилка при збереженні даних.")
		bot.Send(msg)
		return
	}

	log.Printf("✅ Успішно збережено %d об'єктів у %s", len(properties), outputPath)
	msg = tgbotapi.NewMessage(message.Chat.ID, fmt.Sprintf("✅ Збережено %d об'єктів!", len(properties)))
	bot.Send(msg)
}
