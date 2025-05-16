// package main

// import (
// 	"log"
// 	"os"

// 	"github.com/ElRealisto/RieltorGenie/internal/bot"
// 	"github.com/ElRealisto/RieltorGenie/internal/users"
// 	"github.com/joho/godotenv"
// )

// func main() {
// 	err := godotenv.Load()
// 	if err != nil {
// 		log.Fatal("Помилка завантаження .env файлу")
// 	}

// 	token := os.Getenv("TELEGRAM_BOT_TOKEN")
// 	if token == "" {
// 		log.Fatal("TELEGRAM_BOT_TOKEN не встановлений")
// 	}

// 	users.InitDefaultUsers()

// 	b, err := bot.New(token)
// 	if err != nil {
// 		log.Fatalf("Помилка ініціалізації бота: %v", err)
// 	}

//		b.Start()
//	}
// package main

// import (
// 	"log"
// 	"path/filepath"

// 	"github.com/ElRealisto/RieltorGenie/internal/objects"
// )

// func main() {
// 	properties, err := objects.ParseAllCategories()
// 	if err != nil {
// 		log.Fatalf("Помилка при парсингу об'єктів: %v", err)
// 	}

// 	outputPath := filepath.Join("internal", "objects", "parsed_properties.json")

// 	err = objects.SavePropertiesToFile(properties, outputPath)
// 	if err != nil {
// 		log.Fatalf("Помилка при збереженні у файл: %v", err)
// 	}

//		log.Printf("✅ Успішно збережено %d об'єктів у %s", len(properties), outputPath)
//	}
package main

import (
	"log"
	"os"
	"path/filepath"

	"github.com/ElRealisto/RieltorGenie/internal/objects"
	"github.com/ElRealisto/RieltorGenie/internal/users"
	"github.com/joho/godotenv"
)

func main() {
	// Завантаження змінних середовища з .env файлу
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Помилка завантаження .env файлу")
	}

	// Отримання токену Telegram бота з змінних середовища
	token := os.Getenv("TELEGRAM_BOT_TOKEN")
	if token == "" {
		log.Fatal("TELEGRAM_BOT_TOKEN не встановлений")
	}

	// Ініціалізація користувачів
	users.InitDefaultUsers()

	// Отримання списку всіх користувачів
	allUsers := users.GetAll()

	// Парсинг об'єктів для всіх рієлторів
	for _, u := range allUsers {
		if u.Role != users.RealtorRole {
			continue
		}

		properties, err := objects.ParseRealtorProfile(u.ProfileURL)
		if err != nil {
			log.Printf("❌ Помилка парсингу профілю %s: %v", u.ProfileURL, err)
			continue
		}

		outputPath := filepath.Join("internal", "objects", "parsed_properties.json")
		err = objects.SavePropertiesToFile(properties, outputPath)
		if err != nil {
			log.Printf("❌ Помилка збереження JSON: %v", err)
			continue
		}

		log.Printf("✅ Успішно збережено %d об'єктів у %s", len(properties), outputPath)

		// Запуск автопарсингу для поточного рієлтора
		objects.StartAutoParsing(u.ProfileURL, outputPath)

		// Для тестування з одним рієлтором можна зупинити цикл
		break
	}

	// Запобігання завершенню програми, щоб автопарсинг продовжував працювати
	select {}
}
