package objects

import (
	"log"
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

// runScheduledParsing виконує сам парсинг і збереження
func runScheduledParsing(profileURL, filename string) {
	log.Printf("🔍 Починаємо автопарсинг %s...", profileURL)

	results, err := ParseRealtorProfile(profileURL)
	if err != nil {
		log.Println("❌ Помилка при парсингу:", err)
		return
	}

	err = SavePropertiesToFile(results, filename)
	if err != nil {
		log.Println("❌ Помилка при збереженні:", err)
		return
	}

	log.Printf("✅ Збережено %d об'єктів у файл %s", len(results), filename)
}
