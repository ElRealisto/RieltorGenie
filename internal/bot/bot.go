package bot

import (
	"fmt"
	"log"

	"github.com/ElRealisto/RieltorGenie/internal/users"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type Bot struct {
	api *tgbotapi.BotAPI
}

func New(token string) (*Bot, error) {
	api, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		return nil, err
	}
	api.Debug = true
	return &Bot{api: api}, nil
}

func (b *Bot) Start() {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates := b.api.GetUpdatesChan(u)

	for update := range updates {
		if update.Message == nil && update.CallbackQuery == nil {
			continue
		}

		if update.CallbackQuery != nil {
			HandleUpdate(b.api, update)
			continue
		}

		senderID := update.Message.Chat.ID
		matchedUser := users.FindByTelegramID(senderID)

		if matchedUser == nil {
			b.send(senderID, "Вибач, я тебе не знаю.")
			continue
		}

		// Відправка вітального повідомлення для ріелтора
		if matchedUser.Role == users.RealtorRole {
			b.sendWithInlineButtons(senderID, fmt.Sprintf("Вітаю тебе, о %s! Я твій вірний джин! 🔮", matchedUser.Name))
			continue
		}

		if matchedUser.Role == users.AdminRole && update.Message.Text == "/test" {
			b.send(senderID, users.PrintDebugInfo(users.GetAll()))
			continue
		}

		HandleUpdate(b.api, update)
	}
}

func (b *Bot) send(chatID int64, message string) {
	msg := tgbotapi.NewMessage(chatID, message)
	if _, err := b.api.Send(msg); err != nil {
		log.Printf("Помилка надсилання повідомлення: %v", err)
	}
}

func (b *Bot) sendWithInlineButtons(chatID int64, message string) {
	msg := tgbotapi.NewMessage(chatID, message)

	// Створення вбудованої клавіатури з кнопками
	inlineKeyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Завантажити об'єкти", "load_objects"),
			tgbotapi.NewInlineKeyboardButtonData("Оновити об'єкти", "update_objects"),
		),
	)

	msg.ReplyMarkup = inlineKeyboard

	if _, err := b.api.Send(msg); err != nil {
		log.Printf("Помилка надсилання повідомлення: %v", err)
	}
}

func (b *Bot) SendMarkdownMessage(chatID int64, text string) error {
	msg := tgbotapi.NewMessage(chatID, text)
	msg.ParseMode = "Markdown"
	_, err := b.api.Send(msg)
	return err
}
