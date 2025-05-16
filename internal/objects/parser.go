package objects

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

const baseURL = "https://rieltor.ua"

type Property struct {
	Title        string `json:"title"`
	Price        string `json:"price"`
	Link         string `json:"link"`
	Category     string `json:"category"`
	Rooms        string `json:"rooms"`
	Area         string `json:"area"`
	FloorDetails string `json:"floorDetails"`
}

type House struct {
	Title        string `json:"title"`
	Price        string `json:"price"`
	Link         string `json:"link"`
	Category     string `json:"category"`
	Area         string `json:"area"`
	FloorDetails string `json:"floorDetails"`
	LandPlot     string `json:"landPlot,omitempty"`
}

// cleanTitle очищає заголовок від зайвих символів
func cleanTitle(raw string) string {
	cleaned := strings.ReplaceAll(raw, "\n", " ")
	cleaned = strings.ReplaceAll(cleaned, "\t", " ")
	return strings.Join(strings.Fields(cleaned), " ")
}

// findCategoryFromSlug повертає PropertyCategory за частиною URL
func findCategoryFromSlug(slug string) *PropertyCategory {
	for _, cat := range PropertyCategories {
		if cat.RelativePath == slug {
			return &cat
		}
	}
	return nil
}

// createProperty створює Property зі спанів
func createProperty(title, price, link, category string, spans *goquery.Selection) Property {
	return Property{
		Title:        title,
		Price:        price,
		Link:         link,
		Category:     category,
		Rooms:        strings.TrimSpace(spans.Eq(0).Text()),
		Area:         strings.TrimSpace(spans.Eq(1).Text()),
		FloorDetails: strings.TrimSpace(spans.Eq(2).Text()),
	}
}

// createHouse створює House зі спанів
func createHouse(title, price, link, category string, spans *goquery.Selection) House {
	return House{
		Title:        title,
		Price:        price,
		Link:         link,
		Category:     category,
		Area:         strings.TrimSpace(spans.Eq(0).Text()),
		FloorDetails: strings.TrimSpace(spans.Eq(1).Text()),
		LandPlot:     strings.TrimSpace(spans.Eq(2).Text()),
	}
}

// ParseRealtorProfile парсить об’єкти з профілю рієлтора
func ParseRealtorProfile(profileURL string) ([]any, error) {
	var results []any

	res, err := http.Get(profileURL)
	if err != nil {
		return nil, fmt.Errorf("не вдалося отримати сторінку: %w", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("неочікуваний статус-код: %d", res.StatusCode)
	}

	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		return nil, fmt.Errorf("помилка при створенні документа: %w", err)
	}

	// Визначення кількості активних оголошень
	countText := doc.Find(".user_item_activity_text:contains('Активних оголошень')").
		SiblingsFiltered(".user_item_activity_number").
		First().
		Text()

	activeCount, err := strconv.Atoi(strings.TrimSpace(countText))
	if err != nil {
		fmt.Println("⚠️ Не вдалося визначити кількість активних оголошень, парсимо все")
		activeCount = -1
	}

	type categoryInfo struct {
		Name  string
		Count int
	}
	var categories []categoryInfo

	doc.Find("a.agency_title_link_m").Each(func(i int, s *goquery.Selection) {
		href, exists := s.Attr("href")
		if !exists {
			return
		}
		parts := strings.Split(strings.Trim(href, "/"), "/")
		slug := parts[len(parts)-1]

		countParts := strings.Split(s.Text(), "–")
		if len(countParts) == 2 {
			countStr := strings.TrimSpace(countParts[1])
			if count, err := strconv.Atoi(countStr); err == nil {
				categories = append(categories, categoryInfo{Name: slug, Count: count})
			}
		}
	})

	// Основний цикл парсингу об'єктів
	objIndex := 0
	doc.Find(".catalog-card").EachWithBreak(func(i int, s *goquery.Selection) bool {
		if activeCount >= 0 && objIndex >= activeCount {
			return false
		}

		price := strings.TrimSpace(s.Find(".catalog-card-price-title").Text())
		title := cleanTitle(s.Find("h2").Text())
		link, _ := s.Find("a.catalog-card-media").Attr("href")
		if strings.HasPrefix(link, "/") {
			link = baseURL + link
		}

		// Визначення категорії на основі index
		objCategory := "other"
		localIndex := objIndex
		for _, cat := range categories {
			if localIndex < cat.Count {
				objCategory = cat.Name
				break
			}
			localIndex -= cat.Count
		}

		// Аналіз даних об'єкта
		spans := s.Find("div.catalog-card-details-row span")

		if strings.Contains(objCategory, "houses") && spans.Length() >= 3 {
			house := createHouse(title, price, link, objCategory, spans)
			results = append(results, house)
		} else if spans.Length() >= 3 {
			prop := createProperty(title, price, link, objCategory, spans)
			results = append(results, prop)
		}

		objIndex++
		time.Sleep(500 * time.Millisecond)
		return true
	})

	return results, nil
}

// SavePropertiesToFile зберігає результати в JSON
func SavePropertiesToFile(data []any, filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("не вдалося створити файл: %w", err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(data); err != nil {
		return fmt.Errorf("не вдалося записати JSON: %w", err)
	}

	return nil
}
