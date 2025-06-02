// // package main

// // import (
// // 	"log"
// // 	"os"

// // 	"github.com/ElRealisto/RieltorGenie/internal/bot"
// // 	"github.com/ElRealisto/RieltorGenie/internal/users"
// // 	"github.com/joho/godotenv"
// // )

// // func main() {
// // 	err := godotenv.Load()
// // 	if err != nil {
// // 		log.Fatal("Помилка завантаження .env файлу")
// // 	}

// // 	token := os.Getenv("TELEGRAM_BOT_TOKEN")
// // 	if token == "" {
// // 		log.Fatal("TELEGRAM_BOT_TOKEN не встановлений")
// // 	}

// // 	users.InitDefaultUsers()

// // 	b, err := bot.New(token)
// // 	if err != nil {
// // 		log.Fatalf("Помилка ініціалізації бота: %v", err)
// // 	}

// //		b.Start()
// //	}
// // package main

// // import (
// // 	"log"
// // 	"path/filepath"

// // 	"github.com/ElRealisto/RieltorGenie/internal/objects"
// // )

// // func main() {
// // 	properties, err := objects.ParseAllCategories()
// // 	if err != nil {
// // 		log.Fatalf("Помилка при парсингу об'єктів: %v", err)
// // 	}

// // 	outputPath := filepath.Join("internal", "objects", "parsed_objects.json")

// // 	err = objects.SavePropertiesToFile(properties, outputPath)
// // 	if err != nil {
// // 		log.Fatalf("Помилка при збереженні у файл: %v", err)
// // 	}

// //		log.Printf("✅ Успішно збережено %d об'єктів у %s", len(properties), outputPath)
// //	}
// package main

// import (
// 	"log"
// 	"os"
// 	"path/filepath"
// 	"strings"

// 	"github.com/ElRealisto/RieltorGenie/internal/bot"
// 	"github.com/ElRealisto/RieltorGenie/internal/objects"
// 	"github.com/ElRealisto/RieltorGenie/internal/users"
// 	"github.com/joho/godotenv"
// )

// func main() {
// 	// Завантаження змінних середовища з .env файлу
// 	err := godotenv.Load()
// 	if err != nil {
// 		log.Fatal("Помилка завантаження .env файлу")
// 	}

// 	// Отримання токену Telegram бота з змінних середовища
// 	token := os.Getenv("TELEGRAM_BOT_TOKEN")
// 	if token == "" {
// 		log.Fatal("TELEGRAM_BOT_TOKEN не встановлений")
// 	}

// 	// Ініціалізація користувачів
// 	users.InitDefaultUsers()

// 	// Отримання списку всіх користувачів
// 	allUsers := users.GetAll()

// 	// Створення екземпляра бота
// 	b, err := bot.New(token)
// 	if err != nil {
// 		log.Fatalf("Помилка створення бота: %v", err)
// 	}

// 	// Запуск Telegram-бота в окремій горутині
// 	go b.Start()

// 	// Запуск моніторингу в окремій горутині
// 	go objects.StartMonitoring(allUsers)

// 	for _, u := range allUsers {
// 		if u.Role != users.RealtorRole {
// 			continue
// 		}

// 		// Парсинг об'єктів рієлтора
// 		properties, err := objects.ParseRealtorProfile(u.ProfileURL)
// 		if err != nil {
// 			log.Printf("❌ Помилка парсингу профілю %s: %v", u.ProfileURL, err)
// 			continue
// 		}

// 		// Збереження загального списку у parsed_objects.json
// 		internalPath := filepath.Join("internal", "objects", "parsed_objects.json")
// 		err = objects.SavePropertiesToFile(properties, internalPath)
// 		if err != nil {
// 			log.Printf("❌ Помилка збереження JSON: %v", err)
// 			continue
// 		}
// 		log.Printf("✅ Успішно збережено %d об'єктів у %s", len(properties), internalPath)

// 		// Збереження по категоріях у data/
// 		profileSlug := strings.ReplaceAll(strings.TrimPrefix(u.ProfileURL, "https://"), "/", "_")
// 		baseDir := filepath.Join("data", profileSlug)

// 		err = objects.SaveObjectsByCategory(properties, baseDir)
// 		if err != nil {
// 			log.Printf("❌ Помилка збереження об'єктів по категоріях: %v", err)
// 			continue
// 		}
// 		log.Printf("📦 Успішно збережено об'єкти у %s", baseDir)

// 		// Запуск автопарсингу
// 		objects.StartAutoParsing(u.ProfileURL, internalPath)

// 		// Для тестування — вихід після першого рієлтора
// 		break
// 	}

//		select {}
//	}
// package main

// import (
// 	"log"
// 	"os"
// 	"os/exec"
// 	"path/filepath"
// 	"strings"

// 	"github.com/ElRealisto/RieltorGenie/internal/bot"
// 	"github.com/ElRealisto/RieltorGenie/internal/objects"
// 	"github.com/ElRealisto/RieltorGenie/internal/users"
// 	"github.com/joho/godotenv"
// )

// func main() {
// 	// Завантаження змінних середовища з .env файлу
// 	err := godotenv.Load()
// 	if err != nil {
// 		log.Fatal("❌ Помилка завантаження .env файлу")
// 	}

// 	// Отримання токену Telegram бота з змінних середовища
// 	token := os.Getenv("TELEGRAM_BOT_TOKEN")
// 	if token == "" {
// 		log.Fatal("❌ TELEGRAM_BOT_TOKEN не встановлений")
// 	}

// 	// Ініціалізація користувачів
// 	users.InitDefaultUsers()

// 	// Отримання списку всіх користувачів
// 	allUsers := users.GetAll()

// 	// Створення екземпляра бота
// 	b, err := bot.New(token)
// 	if err != nil {
// 		log.Fatalf("❌ Помилка створення бота: %v", err)
// 	}

// 	// Запуск Telegram-бота в окремій горутині
// 	go b.Start()

// 	for _, u := range allUsers {
// 		if u.Role != users.RealtorRole {
// 			continue
// 		}

// 		// Шлях до parsed_objects.json
// 		internalPath := filepath.Join("internal", "objects", "parsed_objects.json")

// 		var properties []objects.Property

// 		// Якщо файл існує — завантажити
// 		if _, err := os.Stat(internalPath); err == nil {
// 			var loaded []objects.Property
// 			loaded, err = objects.LoadParsedObjects()
// 			if err != nil {
// 				log.Printf("❌ Помилка завантаження parsed_objects.json: %v", err)
// 				continue
// 			}
// 			properties = loaded
// 			log.Printf("📂 Завантажено %d об'єктів із %s", len(properties), internalPath)
// 		} else {
// 			// Якщо файл не існує — парсимо профіль
// 			parsed, err := objects.ParseRealtorProfile(u.ProfileURL)
// 			if err != nil {
// 				log.Printf("❌ Помилка парсингу профілю %s: %v", u.ProfileURL, err)
// 				continue
// 			}

// 			// Приводимо []any до []objects.Property
// 			for _, item := range parsed {
// 				prop, ok := item.(objects.Property)
// 				if !ok {
// 					log.Println("❌ Помилка приведення типу до Property")
// 					continue
// 				}
// 				properties = append(properties, prop)
// 			}

// 			// Зберігаємо об'єкти в parsed_objects.json
// 			anyProps := make([]any, len(properties))
// 			for i, p := range properties {
// 				anyProps[i] = p
// 			}

// 			err = objects.SavePropertiesToFile(anyProps, internalPath)
// 			if err != nil {
// 				log.Printf("❌ Помилка збереження JSON: %v", err)
// 				continue
// 			}
// 			log.Printf("✅ Збережено %d об'єктів у %s", len(properties), internalPath)

// 			// Зберігаємо по категоріях
// 			profileSlug := strings.ReplaceAll(strings.TrimPrefix(u.ProfileURL, "https://"), "/", "_")
// 			baseDir := filepath.Join("data", profileSlug)

// 			err = objects.SaveObjectsByCategory(anyProps, baseDir)
// 			if err != nil {
// 				log.Printf("❌ Помилка збереження об'єктів по категоріях: %v", err)
// 				continue
// 			}
// 			log.Printf("📦 Об'єкти збережено у %s", baseDir)
// 		}

// 		// Запуск автопарсингу
// 		objects.StartAutoParsing(u.ProfileURL, internalPath)

// 		// 🔥 УСІГДА генеруємо search_URLs.json після завантаження або парсингу
// 		if err := runGenerateStreetURL(); err != nil {
// 			log.Printf("❌ Помилка запуску скрипта generateStreetURL.js: %v", err)
// 			continue
// 		}
// 		log.Println("🌐 Скрипт generateStreetURL.js успішно завершено")

// 		notifyFunc := func(realtor users.User, message string) error {
// 			return b.SendMarkdownMessage(realtor.TelegramID, message)
// 		}

// 		// Запуск моніторингу об'єктів з файлу search_URLs.json
// 		go objects.StartMonitoring(allUsers, notifyFunc)

// 		// Автопарсинг з логікою моніторингу
// 		// go objects.StartAutoParsingWithMonitoring(u.ProfileURL, internalPath, runGenerateStreetURL)

// 		// Зупиняємось на першому рієлторі
// 		break
// 	}

// 	// Нескінченний блокуючий select, щоб не завершувався main
// 	select {}
// }

// // runGenerateStreetURL запускає Node.js скрипт для генерації search_URLs.json
//
//	func runGenerateStreetURL() error {
//		cmd := exec.Command("node", "scripts/generateStreetURL.js")
//		cmd.Stdout = os.Stdout
//		cmd.Stderr = os.Stderr
//		return cmd.Run()
//	}
package main

import (
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/ElRealisto/RieltorGenie/internal/bot"
	"github.com/ElRealisto/RieltorGenie/internal/objects"
	"github.com/ElRealisto/RieltorGenie/internal/users"
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("❌ Помилка завантаження .env файлу")
	}

	token := os.Getenv("TELEGRAM_BOT_TOKEN")
	if token == "" {
		log.Fatal("❌ TELEGRAM_BOT_TOKEN не встановлений")
	}

	users.InitDefaultUsers()
	allUsers := users.GetAll()

	b, err := bot.New(token)
	if err != nil {
		log.Fatalf("❌ Помилка створення бота: %v", err)
	}

	go b.Start()

	for _, u := range allUsers {
		if u.Role != users.RealtorRole {
			continue
		}

		internalPath := filepath.Join("internal", "objects", "parsed_objects.json")

		if _, err := os.Stat(internalPath); err == nil {
			loaded, err := objects.LoadParsedObjects()
			if err != nil {
				log.Printf("❌ Помилка завантаження parsed_objects.json: %v", err)
				continue
			}
			log.Printf("📂 Завантажено %d об'єктів із %s", len(loaded), internalPath)
		} else {
			parsed, err := objects.ParseRealtorProfile(u.ProfileURL)
			if err != nil {
				log.Printf("❌ Помилка парсингу профілю %s: %v", u.ProfileURL, err)
				continue
			}

			err = objects.SavePropertiesToFile(parsed, internalPath)
			if err != nil {
				log.Printf("❌ Помилка збереження JSON: %v", err)
				continue
			}
			log.Printf("✅ Збережено %d об'єктів у %s", len(parsed), internalPath)

			profileSlug := strings.ReplaceAll(strings.TrimPrefix(u.ProfileURL, "https://"), "/", "_")
			baseDir := filepath.Join("data", profileSlug)

			err = objects.SaveObjectsByCategory(parsed, baseDir)
			if err != nil {
				log.Printf("❌ Помилка збереження об'єктів по категоріях: %v", err)
				continue
			}
			log.Printf("📦 Об'єкти збережено у %s", baseDir)
		}

		objects.StartAutoParsing(u.ProfileURL, internalPath)

		if err := runGenerateStreetURL(); err != nil {
			log.Printf("❌ Помилка запуску скрипта generateStreetURL.js: %v", err)
			continue
		}
		log.Println("🌐 Скрипт generateStreetURL.js успішно завершено")

		notifyFunc := func(realtor users.User, message string) error {
			return b.SendMarkdownMessage(realtor.TelegramID, message)
		}

		go objects.StartMonitoring(allUsers, notifyFunc)

		break
	}

	select {}
}

func runGenerateStreetURL() error {
	cmd := exec.Command("node", "scripts/generateStreetURL.js")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
