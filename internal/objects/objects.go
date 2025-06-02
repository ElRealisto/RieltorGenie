package objects

import (
	"fmt"
)

// Category тип: продаж чи оренда
type Category string

// PropertyType тип об'єкта: квартира, кімната тощо
type PropertyType string

const (
	CategorySale Category = "sale"
	CategoryRent Category = "rent"

	PropertyTypeFlat       PropertyType = "flats"
	PropertyTypeRoom       PropertyType = "rooms"
	PropertyTypeHouse      PropertyType = "houses"
	PropertyTypeCommercial PropertyType = "commercials"
	PropertyTypeLand       PropertyType = "areas"
)

// PropertyCategory описує категорію та тип нерухомості
type PropertyCategory struct {
	Category     Category
	Type         PropertyType
	DisplayUkr   string // Наприклад: "Комерційна нерухомість"
	RelativePath string // Наприклад: "commercials-sale"
}

// PropertyCategories — список усіх категорій і типів
var PropertyCategories = []PropertyCategory{
	{CategorySale, PropertyTypeFlat, "Квартира", "flats-sale"},
	{CategorySale, PropertyTypeRoom, "Кімната", "rooms-sale"},
	{CategorySale, PropertyTypeHouse, "Будинок", "houses-sale"},
	{CategorySale, PropertyTypeCommercial, "Комерційна нерухомість", "commercials-sale"},
	{CategorySale, PropertyTypeLand, "Земля", "areas-sale"},

	{CategoryRent, PropertyTypeFlat, "Квартира", "flats-rent"},
	{CategoryRent, PropertyTypeRoom, "Кімната", "rooms-rent"},
	{CategoryRent, PropertyTypeHouse, "Будинок", "houses-rent"},
	{CategoryRent, PropertyTypeCommercial, "Комерційна нерухомість", "commercials-rent"},
	{CategoryRent, PropertyTypeLand, "Земля", "areas-rent"},
}

// GenerateURL формує повне посилання на сторінку з об'єктами
func GenerateURL(domain string, pc PropertyCategory) string {
	return fmt.Sprintf("https://%s/%s/", domain, pc.RelativePath)
}

// StartAutoParsingWithMonitoring запускає автопарсинг і моніторинг у циклі
// func StartAutoParsingWithMonitoring(profileURL, parsedObjectsPath string, onUpdate func() error) {
// 	for {
// 		log.Println("🔁 Починаємо автопарсинг профілю:", profileURL)

// 		// 1. Парсинг профілю
// 		parsed, err := ParseRealtorProfile(profileURL)
// 		if err != nil {
// 			log.Printf("❌ Помилка автопарсингу: %v", err)
// 			time.Sleep(10 * time.Minute)
// 			continue
// 		}

// 		// 2. Приведення типів: []any → []Property
// 		var properties []Property
// 		for _, item := range parsed {
// 			prop, ok := item.(Property)
// 			if !ok {
// 				log.Println("⚠️ Неправильний тип при приведенні до Property")
// 				continue
// 			}
// 			properties = append(properties, prop)
// 		}

// 		// 3. Збереження оновлених об'єктів у файл
// 		anyProps := make([]any, len(properties))
// 		for i, p := range properties {
// 			anyProps[i] = p
// 		}

// 		err = SavePropertiesToFile(anyProps, parsedObjectsPath)
// 		if err != nil {
// 			log.Printf("❌ Помилка збереження об'єктів: %v", err)
// 		} else {
// 			log.Printf("✅ Збережено %d об'єктів у %s", len(properties), parsedObjectsPath)
// 		}

// 		// 4. Оновити search_URLs.json через передану функцію
// 		if err := onUpdate(); err != nil {
// 			log.Printf("❌ Помилка оновлення через onUpdate: %v", err)
// 		}

// 		// 5. Затримка до наступного запуску
// 		time.Sleep(30 * time.Minute) // або інтервал, який хочеш
// 	}
// }
