package objects

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/ElRealisto/RieltorGenie/internal/users"
	"github.com/chromedp/chromedp"
)

type SearchResult struct {
	Title    string `json:"title"`
	Price    string `json:"price"`
	Link     string `json:"link"`
	Position int    `json:"position"`
	PostedBy string `json:"posted_by"`
}

type SearchResults struct {
	Results []SearchResult `json:"search_results"`
}

type SearchURL struct {
	Title string `json:"title"`
	URL   string `json:"url"`
}

// MonitoredObject — легкий аналог Property без циклічного імпорту
type MonitoredObject struct {
	Title string `json:"title"`
	Link  string `json:"link"`
}

func (m MonitoredObject) URL() string {
	return m.Link
}

func StartMonitoring(realtors []users.User, notify func(realtor users.User, message string) error) {
	log.Println("🚀 Моніторинг позицій об'єктів запущено")

	go func() {
		for {
			log.Println("🔁 Нова ітерація моніторингу")

			for _, realtor := range realtors {
				if realtor.Role != users.RealtorRole {
					continue
				}

				log.Printf("👤 Обробка ріелтора %s (%s)", realtor.Name, realtor.ProfileURL)

				objects := GetObjectsForRealtor(realtor.ProfileURL)
				log.Printf("📂 Отримання об'єктів завершено, знайдено %d об'єктів", len(objects))

				for i, obj := range objects {
					log.Printf("🔍 Перевірка об'єкта %d (%s) ріелтора %s", i+1, obj.Title, realtor.Name)

					isFirst, topRealtor := CheckObjectPosition(obj, realtor)
					log.Println("✅ Перевірка завершена")

					if !isFirst {
						log.Printf("⚠️ Рієлтор %s НЕ перший для об'єкта: %s", realtor.Name, obj.Title)

						message := fmt.Sprintf(
							"⚠️ О, світ очей моїх! Ваш об'єкт *%s* рекламує першим якась падла: *%s*\nПосилання: %s",
							obj.Title, topRealtor, obj.URL(),
						)

						err := notify(realtor, message)
						if err != nil {
							log.Printf("❌ Помилка надсилання повідомлення ріелтору %s: %v", realtor.Name, err)
						} else {
							log.Printf("📨 Повідомлення надіслано ріелтору %s", realtor.Name)
						}
					} else {
						log.Printf("✅ Наш ріелтор %s перший для об'єкта %s", realtor.Name, obj.Title)
					}
					log.Println("✅ Перевірка об'єкта завершена, чекаємо 15 секунд...")
					time.Sleep(15 * time.Second)
				}
			}

			log.Println("🔁 Ітерація моніторингу завершена. Чекаємо 60 секунд до наступної...")
			time.Sleep(60 * time.Second)
		}
	}()
}

func GetObjectsForRealtor(profileURL string) []MonitoredObject {
	all, err := LoadParsedObjects()
	if err != nil {
		log.Printf("❌ Помилка завантаження обʼєктів: %v", err)
		return nil
	}

	var monitored []MonitoredObject
	for _, p := range all {
		switch obj := p.(type) {
		case Property:
			monitored = append(monitored, MonitoredObject{
				Title: obj.Title,
				Link:  obj.Link,
			})
		case House:
			monitored = append(monitored, MonitoredObject{
				Title: obj.Title,
				Link:  obj.Link,
			})
		default:
			log.Printf("⚠️ Невідомий тип об'єкта: %T", p)
		}
	}

	return monitored
}

func CheckObjectPosition(obj MonitoredObject, realtor users.User) (bool, string) {
	urlMap, err := loadSearchURLs()
	if err != nil {
		log.Printf("❌ Помилка завантаження search_URLs.json: %v", err)
		return false, ""
	}

	searchURL, ok := urlMap[obj.Title]
	if !ok {
		log.Printf("⚠️ Немає пошукового URL для обʼєкта: %s", obj.Title)
		return false, ""
	}

	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("headless", true),
		chromedp.Flag("disable-gpu", true),
		chromedp.Flag("no-sandbox", true),
		chromedp.Flag("disable-dev-shm-usage", true),
	)
	allocCtx, cancelAlloc := chromedp.NewExecAllocator(context.Background(), opts...)
	defer cancelAlloc()

	ctxBase, cancelBase := chromedp.NewContext(allocCtx)
	defer cancelBase()

	ctx, cancel := context.WithTimeout(ctxBase, 15*time.Second)
	defer cancel()

	log.Printf("🕵️‍♀️ Перехід за URL: %s", searchURL)

	var countText string
	err = chromedp.Run(ctx,
		chromedp.Navigate(searchURL),
		chromedp.Sleep(2*time.Second),
		chromedp.Evaluate(`document.querySelector('span[data-listing-count]')?.textContent`, &countText),
	)
	if err != nil {
		log.Printf("❌ Помилка під час завантаження сторінки: %v", err)
		return false, ""
	}

	log.Printf("📊 Всього оголошень: '%s'", countText)

	if countText == "" {
		var html string
		_ = chromedp.Run(ctx, chromedp.OuterHTML("html", &html))
		fmt.Println("⚠️ DOM HTML (обрізаний):\n", html[:1000])
		log.Println("❌ Не знайдено елемент 'span[data-listing-count]'")
		return false, ""
	}

	count := extractNumber(countText)
	if count == 0 {
		fmt.Printf("❌ Обʼєкт '%s' не знайдено серед результатів\n", obj.Title)
		return false, ""
	}

	var titles, authors []string

	titlesJS := fmt.Sprintf(`Array.from(document.querySelectorAll('.catalog-card-address')).slice(0, %d).map(e => e.textContent.trim())`, count)
	authorsJS := fmt.Sprintf(`Array.from(document.querySelectorAll('button.catalog-card-author-title')).slice(0, %d).map(e => e.textContent.trim())`, count)

	err = chromedp.Run(ctx,
		chromedp.Evaluate(titlesJS, &titles),
		chromedp.Evaluate(authorsJS, &authors),
	)
	if err != nil {
		log.Printf("❌ Помилка під час обробки результатів: %v", err)
		return false, ""
	}

	if len(titles) == 0 || len(authors) == 0 {
		fmt.Printf("❌ Обʼєкт '%s' не знайдено серед результатів\n", obj.Title)
		return false, ""
	}

	if !strings.EqualFold(titles[0], obj.Title) {
		fmt.Printf("⚠️ Обʼєкт '%s' знайдено не першим: перший у списку — '%s'\n", obj.Title, titles[0])
		return false, authors[0]
	}

	if !strings.EqualFold(authors[0], realtor.Name) {
		fmt.Printf("⚠️ Обʼєкт '%s' знайдено, але першим його розміщує інший ріелтор: %s\n", obj.Title, authors[0])
		return false, authors[0]
	}

	fmt.Printf("✅ Обʼєкт '%s' знайдено, першим розміщує наш ріелтор: %s\n", obj.Title, authors[0])
	return true, authors[0]
}

func extractNumber(text string) int {
	num := 0
	for _, r := range text {
		if r >= '0' && r <= '9' {
			num = num*10 + int(r-'0')
		}
	}
	return num
}

func loadSearchURLs() (map[string]string, error) {
	data, err := os.ReadFile("internal/objects/search_URLs.json")
	if err != nil {
		return nil, err
	}

	var urlList []SearchURL
	err = json.Unmarshal(data, &urlList)
	if err != nil {
		return nil, err
	}

	urlMap := make(map[string]string)
	for _, item := range urlList {
		urlMap[item.Title] = item.URL
	}

	return urlMap, nil
}
