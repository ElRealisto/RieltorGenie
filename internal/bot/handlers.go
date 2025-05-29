package bot

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/ElRealisto/RieltorGenie/internal/objects"
	"github.com/ElRealisto/RieltorGenie/internal/users"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// Property описує об'єкт нерухомості
type Property struct {
	Title        string `json:"title"`
	Price        string `json:"price"`
	Link         string `json:"link"`
	Category     string `json:"category"`
	Rooms        string `json:"rooms"`
	Area         string `json:"area"`
	FloorDetails string `json:"floorDetails"`
	LandPlot     string `json:"landPlot,omitempty"` // Додано для будинків
}

// HandleUpdate обробляє вхідні повідомлення Telegram
func HandleUpdate(bot *tgbotapi.BotAPI, update tgbotapi.Update) {
	if update.Message == nil && update.CallbackQuery == nil {
		return
	}

	// Обробка вбудованих кнопок
	if update.CallbackQuery != nil {
		callback := tgbotapi.NewCallback(update.CallbackQuery.ID, "")
		if _, err := bot.Request(callback); err != nil {
			log.Printf("Помилка підтвердження callback: %v", err)
		}

		switch update.CallbackQuery.Data {
		case "load_objects":
			// Логіка для завантаження об'єктів з файлу
			handleLoadObjects(bot, update.CallbackQuery.Message)
		case "update_objects":
			// Логіка для оновлення об'єктів
			handleParseCommand(bot, update.CallbackQuery.Message)
		}
		return
	}

	// Обробка текстових повідомлень
	switch update.Message.Text {
	case "/парсити":
		handleParseCommand(bot, update.Message)
	default:
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "🤖 Невідома команда. Спробуйте /парсити")
		bot.Send(msg)
	}
}

// handleLoadObjects завантажує об'єкти з файлу і відправляє їх користувачу
func handleLoadObjects(bot *tgbotapi.BotAPI, message *tgbotapi.Message) {
	filePath := filepath.Join("internal", "objects", "parsed_objects.json")

	// Читання файлу
	fileContent, err := os.ReadFile(filePath)
	if err != nil {
		log.Printf("Помилка читання файлу: %v", err)
		msg := tgbotapi.NewMessage(message.Chat.ID, "❌ Помилка при читанні файлу.")
		bot.Send(msg)
		return
	}

	// Розпарсити JSON
	var properties []Property
	err = json.Unmarshal(fileContent, &properties)
	if err != nil {
		log.Printf("Помилка розпарсингу JSON: %v", err)
		msg := tgbotapi.NewMessage(message.Chat.ID, "❌ Помилка при розпарсингу даних.")
		bot.Send(msg)
		return
	}

	// Формування повідомлення з об'єктами
	responseMessage := fmt.Sprintf("Наразі Ви рекламуєте %d об'єктів:\n\n", len(properties))
	for i, property := range properties {
		categoryDescription := getCategoryDescription(property.Category)
		responseMessage += fmt.Sprintf("%d. 🏠 Назва: %s\n💰 Ціна: %s\n🔗 Посилання: %s\n🏡 Кімнати: %s\n📏 Площа: %s\n🏢 Поверх: %s\n📌 Категорія: %s\n\n",
			i+1, property.Title, property.Price, property.Link, property.Rooms, property.Area, property.FloorDetails, categoryDescription)
	}

	// Відправка повідомлення з об'єктами
	msg := tgbotapi.NewMessage(message.Chat.ID, responseMessage)
	bot.Send(msg)

	// Відправка кнопок знову
	sendWithInlineButtons(bot, message.Chat.ID, "Оберіть дію:")
}

// sendWithInlineButtons відправляє повідомлення з вбудованою клавіатурою
func sendWithInlineButtons(bot *tgbotapi.BotAPI, chatID int64, message string) {
	msg := tgbotapi.NewMessage(chatID, message)

	// Створення вбудованої клавіатури з кнопками
	inlineKeyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Завантажити об'єкти", "load_objects"),
			tgbotapi.NewInlineKeyboardButtonData("Оновити об'єкти", "update_objects"),
		),
	)

	msg.ReplyMarkup = inlineKeyboard

	if _, err := bot.Send(msg); err != nil {
		log.Printf("Помилка надсилання повідомлення: %v", err)
	}
}

// getCategoryDescription повертає опис категорії
func getCategoryDescription(category string) string {
	switch category {
	case "flats-sale":
		return "Продаж квартири"
	case "flats-rent":
		return "Оренда квартири"
	case "houses-sale":
		return "Продаж будинку"
	case "areas-sale":
		return "Продаж землі"
	case "commercials-sale":
		return "Продаж комерційної нерухомості"
	case "commercials-rent":
		return "Оренда комерційної нерухомості"
	default:
		return "Інше"
	}
}

// handleParseCommand запускає парсинг для рієлтора
func handleParseCommand(bot *tgbotapi.BotAPI, message *tgbotapi.Message) {
	// Переконайтеся, що ви використовуєте правильний Telegram ID
	telegramID := message.Chat.ID
	fmt.Printf("🔍 Пошук користувача з Telegram ID: %d\n", telegramID)

	user := users.FindByTelegramID(telegramID)
	if user == nil {
		msg := tgbotapi.NewMessage(message.Chat.ID, "🚫 Вас не знайдено у базі рієлторів.")
		bot.Send(msg)
		return
	}

	fmt.Printf("🔍 Знайдено користувача: %+v\n", user)

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

	outputPath := filepath.Join("internal", "objects", "parsed_objects.json")
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
