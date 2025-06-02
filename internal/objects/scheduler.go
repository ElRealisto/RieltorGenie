package objects

import (
	"bytes"
	"encoding/json"
	"log"
	"os"
	"os/exec"
	"time"

	"github.com/robfig/cron/v3"
)

// StartAutoParsing запускає автоматичний парсинг за розкладом
func StartAutoParsing(profileURL, filename string) {
	loc, err := time.LoadLocation("Europe/Kyiv")
	if err != nil {
		log.Fatalf("Не вдалося завантажити часову зону: %v", err)
	}

	c := cron.New(
		cron.WithLocation(loc),
		cron.WithSeconds(), // дозволяє задавати cron з секундами
	)

	// Щодня о 00:00
	_, err = c.AddFunc("0 0 0 * * *", func() {
		runScheduledParsing(profileURL, filename)
	})
	if err != nil {
		log.Println("Помилка при додаванні завдання на 00:00:", err)
	}

	// О 11:00 з понеділка по п'ятницю
	_, err = c.AddFunc("0 0 11 * * 1-5", func() {
		runScheduledParsing(profileURL, filename)
	})
	if err != nil {
		log.Println("Помилка при додаванні завдання на 11:00:", err)
	}

	// О 18:00 з понеділка по п'ятницю
	_, err = c.AddFunc("0 0 18 * * 1-5", func() {
		runScheduledParsing(profileURL, filename)
	})
	if err != nil {
		log.Println("Помилка при додаванні завдання на 18:00:", err)
	}

	c.Start()
	log.Println("⏰ Автопарсинг запущено за розкладом...")
}

// runScheduledParsing виконує парсинг, порівняння, збереження і запуск JS при зміні
func runScheduledParsing(profileURL, filename string) {
	log.Printf("🔍 Починаємо автопарсинг %s...", profileURL)

	// Крок 1: Отримати нові об'єкти
	results, err := ParseRealtorProfile(profileURL)
	if err != nil {
		log.Println("❌ Помилка при парсингу:", err)
		return
	}

	// Крок 2: Перевірити, чи відрізняються нові дані від існуючого файлу
	changed, err := hasPropertiesChanged(results, filename)
	if err != nil {
		log.Println("⚠️ Помилка при порівнянні файлів:", err)
		return
	}

	if !changed {
		log.Println("🟡 Дані не змінилися — збереження і оновлення не потрібні.")
		return
	}

	// Крок 3: Зберегти нові об'єкти у файл
	err = SavePropertiesToFile(results, filename)
	if err != nil {
		log.Println("❌ Помилка при збереженні:", err)
		return
	}
	log.Printf("✅ Збережено %d об'єктів у файл %s", len(results), filename)

	// Крок 4: Запустити generateStreetURL.js
	err = runGenerateStreetURL()
	if err != nil {
		log.Println("⚠️ Помилка при запуску generateStreetURL.js:", err)
		return
	}
	log.Println("🔄 Оновлено search_URLs.json через generateStreetURL.js")
}

// hasPropertiesChanged порівнює нові дані з уже збереженими
func hasPropertiesChanged(newData any, filename string) (bool, error) {
	// Переводимо нові дані в JSON
	newJSON, err := json.MarshalIndent(newData, "", "  ")
	if err != nil {
		return false, err
	}

	// Зчитуємо старий файл
	oldJSON, err := os.ReadFile(filename)
	if err != nil {
		// Якщо файл не існує — вважаємо, що дані змінилися
		if os.IsNotExist(err) {
			return true, nil
		}
		return false, err
	}

	// Порівнюємо вміст
	return !bytes.Equal(newJSON, oldJSON), nil
}

// runGenerateStreetURL викликає скрипт generateStreetURL.js
func runGenerateStreetURL() error {
	cmd := exec.Command("node", "scripts/generateStreetURL.js")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
