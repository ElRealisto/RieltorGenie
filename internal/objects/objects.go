package objects

import "fmt"

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
