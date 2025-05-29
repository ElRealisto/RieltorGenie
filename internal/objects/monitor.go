package objects

import (
	"context"
	"fmt"
	"log"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/ElRealisto/RieltorGenie/internal/users"
	"github.com/chromedp/chromedp"
)

// SearchResult описує результат пошуку
type SearchResult struct {
	Title    string `json:"title"`
	Price    string `json:"price"`
	Link     string `json:"link"`
	Position int    `json:"position"`
	PostedBy string `json:"posted_by"`
}

// SearchResults описує список результатів пошуку
type SearchResults struct {
	Results []SearchResult `json:"search_results"`
}

// StartMonitoring запускає нескінченний цикл перевірок позицій об'єктів
func StartMonitoring(realtors []users.User) {
	fmt.Println("🚀 Моніторинг позицій об'єктів запущено")

	go func() {
		for {
			for _, realtor := range realtors {
				if realtor.Role != users.RealtorRole {
					continue
				}

				// Отримати об'єкти цього ріелтора
				objects := GetObjectsForRealtor(realtor.ProfileURL)

				// Перевіряємо всі об'єкти ріелтора
				for i, obj := range objects {
					fmt.Printf("🔍 Перевірка об'єкта %d (%s) ріелтора %s...\n", i+1, obj.Title, realtor.Name)
					CheckObjectPosition(obj, realtor)
					time.Sleep(10 * time.Second) // Пауза між перевірками об'єктів
				}
			}

			// Пауза між циклами (наприклад, 3 хвилини)
			time.Sleep(3 * time.Minute)
		}
	}()
}

// GetObjectsForRealtor — завантажує об'єкти з parsed_objects.json
func GetObjectsForRealtor(profileURL string) []Property {
	all, err := LoadParsedObjects()
	if err != nil {
		log.Printf("❌ Помилка завантаження обʼєктів: %v", err)
		return nil
	}

	// Тимчасово повертаємо всі об'єкти без фільтрації
	return all
}

// CheckObjectPosition — основна логіка перевірки позиції об'єкта
func CheckObjectPosition(obj Property, realtor users.User) {
	ctx, cancel := chromedp.NewContext(context.Background())
	defer cancel()

	var allRealtors []string

	// Формуємо URL для пошуку на основі даних об'єкта
	priceMin, priceMax := calculatePriceRange(obj.Price)
	streetName := url.QueryEscape(obj.Title)
	searchURL := fmt.Sprintf("https://rieltor.ua/%s/%s-rooms/?currency=2&price_min=%d&price_max=%d&radius=20&sort=-default&street_name=%s#15.59/50.435089/30.511846",
		obj.Category,
		strings.TrimSuffix(obj.Rooms, " кімнати"),
		priceMin,
		priceMax,
		streetName)

	fmt.Printf("🔍 Виконання пошуку за URL: %s\n", searchURL)

	// Виконання дій з chromedp
	var actions []chromedp.Action

	actions = append(actions,
		chromedp.Navigate(searchURL),
		chromedp.WaitVisible("div.catalog-items-container", chromedp.ByQuery),
		chromedp.Evaluate(`Array.from(document.querySelectorAll('div.catalog-items-container button.button-link.catalog-card-author-title')).slice(0, 3).map(el => el.innerText)`, &allRealtors),
		chromedp.ActionFunc(func(ctx context.Context) error {
			fmt.Println("🔍 Перелік перших трьох ріелторів на сторінці:")
			for i, realtor := range allRealtors {
				fmt.Printf("%d. %s\n", i+1, realtor)
			}
			return nil
		}),
	)

	err := chromedp.Run(ctx, actions...)
	if err != nil {
		log.Printf("❌ Помилка при перевірці позиції об'єкта: %v", err)
		return
	}

	fmt.Printf("🧪 Перевірка позиції для об'єкта '%s' (%s) ріелтора %s...\n", obj.Title, obj.Link, realtor.Name)

	if len(allRealtors) > 0 {
		if !strings.EqualFold(allRealtors[0], realtor.Name) {
			fmt.Printf("⚠️ Об'єкт '%s' на позиції 1, але його розмістив(ла) інший ріелтор: %s\n",
				obj.Title, allRealtors[0])
			// TODO: Надіслати повідомлення через бот
		} else {
			fmt.Printf("✅ Об'єкт '%s' знайдено на позиції 1, все гаразд\n", obj.Title)
		}
	}
}

// calculatePriceRange обчислює мінімальну і максимальну ціну з урахуванням 15% знижки
func calculatePriceRange(priceStr string) (int, int) {
	re := regexp.MustCompile(`\d+`)
	priceStr = strings.Join(re.FindAllString(priceStr, -1), "")
	priceValue, err := strconv.Atoi(priceStr)
	if err != nil {
		return 0, 0
	}
	minPrice := int(float64(priceValue) * 0.85)
	return minPrice, priceValue
}
