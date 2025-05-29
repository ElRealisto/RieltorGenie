package objects

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
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
	Location     string `json:"location"` // Додано поле Location
	Region       string `json:"region"`   // Додано поле Region
}

type House struct {
	Title        string `json:"title"`
	Price        string `json:"price"`
	Link         string `json:"link"`
	Category     string `json:"category"`
	Area         string `json:"area"`
	FloorDetails string `json:"floorDetails"`
	LandPlot     string `json:"landPlot,omitempty"`
	Location     string `json:"location"` // Додано поле Location
	Region       string `json:"region"`   // Додано поле Region
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
func createProperty(title, price, link, category, region, location string, spans *goquery.Selection) Property {
	return Property{
		Title:        title,
		Price:        price,
		Link:         link,
		Category:     category,
		Rooms:        strings.TrimSpace(spans.Eq(0).Text()),
		Area:         strings.TrimSpace(spans.Eq(1).Text()),
		FloorDetails: strings.TrimSpace(spans.Eq(2).Text()),
		Location:     location, // Додано поле Location
		Region:       region,   // Додано поле Region
	}
}

// createHouse створює House зі спанів
func createHouse(title, price, link, category, region, location string, spans *goquery.Selection) House {
	return House{
		Title:        title,
		Price:        price,
		Link:         link,
		Category:     category,
		Area:         strings.TrimSpace(spans.Eq(0).Text()),
		FloorDetails: strings.TrimSpace(spans.Eq(1).Text()),
		LandPlot:     strings.TrimSpace(spans.Eq(2).Text()),
		Location:     location, // Додано поле Location
		Region:       region,   // Додано поле Region
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
		title := cleanTitle(s.Find(".catalog-card-address").Text())
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

		// Отримання регіону та місця розташування
		region := strings.TrimSpace(s.Find(".catalog-card-region a:first-child").Text())
		location := strings.TrimSpace(s.Find(".catalog-card-region a:last-child").Text())

		// Аналіз даних об'єкта
		spans := s.Find("div.catalog-card-details-row span")

		if strings.Contains(objCategory, "houses") && spans.Length() >= 3 {
			house := createHouse(title, price, link, objCategory, region, location, spans)
			results = append(results, house)
		} else if spans.Length() >= 3 {
			prop := createProperty(title, price, link, objCategory, region, location, spans)
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

func SaveObjectsByCategory(data []any, baseDir string) error {
	// Групуємо по категоріях
	grouped := map[string][]any{}

	for _, v := range data {
		var slug string
		switch obj := v.(type) {
		case Property:
			slug = obj.Category
		case House:
			slug = obj.Category
		default:
			continue
		}
		grouped[slug] = append(grouped[slug], v)
	}

	// Зберігаємо кожну групу
	for slug, items := range grouped {
		pc := findCategoryFromSlug(slug)
		if pc == nil {
			fmt.Printf("⚠️ Невідома категорія: %s\n", slug)
			continue
		}

		dir := filepath.Join(baseDir, string(pc.Category))
		if err := os.MkdirAll(dir, os.ModePerm); err != nil {
			return fmt.Errorf("не вдалося створити директорію %s: %w", dir, err)
		}

		filename := filepath.Join(dir, string(pc.Type)+".json")
		file, err := os.Create(filename) // перезаписуємо файл щоразу
		if err != nil {
			return fmt.Errorf("не вдалося створити файл %s: %w", filename, err)
		}
		defer file.Close()

		encoder := json.NewEncoder(file)
		encoder.SetIndent("", "  ")
		if err := encoder.Encode(items); err != nil {
			return fmt.Errorf("не вдалося записати JSON у файл %s: %w", filename, err)
		}
	}

	return nil
}

var parsedFilePath string

func init() {
	_, filename, _, _ := runtime.Caller(0)
	dir := filepath.Dir(filename)
	parsedFilePath = filepath.Join(dir, "parsed_objects.json")
}

func LoadParsedObjects() ([]Property, error) {
	// Перевірка, чи існує файл
	if _, err := os.Stat(parsedFilePath); os.IsNotExist(err) {
		fmt.Println("⚠️ Файл parsed_objects.json не існує, спарсимо об'єкти...")
		return nil, nil
	}

	file, err := os.Open(parsedFilePath)
	if err != nil {
		return nil, fmt.Errorf("не вдалося відкрити parsed_objects.json: %v", err)
	}
	defer file.Close()

	var props []Property
	err = json.NewDecoder(file).Decode(&props)
	if err != nil {
		return nil, fmt.Errorf("не вдалося розпарсити обʼєкти: %v", err)
	}
	return props, nil
}
