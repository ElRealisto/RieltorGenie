package users

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

type UserRole string

const (
	AdminRole   UserRole = "admin"
	RealtorRole UserRole = "realtor"
)

type User struct {
	TelegramID int64    `json:"telegram_id"`
	ProfileURL string   `json:"profile_url"`
	Name       string   `json:"name"`
	Role       UserRole `json:"role"`
}

var users []User

var usersFile string

func init() { // Файл users.json завжди буде розташований в папці users
	_, filename, _, _ := runtime.Caller(0)
	dir := filepath.Dir(filename)
	usersFile = filepath.Join(dir, "users.json")
}

func LoadUsers() error {
	file, err := os.Open(usersFile)
	if err != nil {
		return fmt.Errorf("не вдалося відкрити файл: %v", err)
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	err = decoder.Decode(&users)
	if err != nil {
		return fmt.Errorf("не вдалося розпакувати дані з файлу: %v", err)
	}

	// Парсимо ім’я лише для ріелторів з порожнім полем Name
	needSave := false
	for i := range users {
		if users[i].Role == RealtorRole && users[i].Name == "" {
			name, err := ParseNameFromProfile(users[i].ProfileURL)
			if err != nil {
				fmt.Printf("Не вдалося парсити ім'я для ріелтора %d: %v\n", users[i].TelegramID, err)
			} else {
				users[i].Name = name
				needSave = true
			}
		}
	}

	// ЗБЕРЕГТИ ОНОВЛЕНІ ІМЕНА
	if needSave {
		return SaveUsers()
	}
	return nil
}

func SaveUsers() error {
	file, err := os.Create(usersFile)
	if err != nil {
		return err
	}
	defer file.Close()

	enc := json.NewEncoder(file)
	enc.SetIndent("", "  ")
	return enc.Encode(users)
}

func GetAll() []User {
	return users
}

func FindByTelegramID(id int64) *User {
	for i := range users {
		if users[i].TelegramID == id {
			fmt.Printf("🔍 Знайдено користувача: %+v\n", users[i])
			return &users[i]
		}
	}
	fmt.Println("❌ Користувача не знайдено.")
	return nil
}

func InitDefaultUsers() {
	if _, err := os.Stat(usersFile); os.IsNotExist(err) {
		users = []User{
			{TelegramID: 5679227412, Role: AdminRole, Name: "Кирило"},
			{TelegramID: 5264545653, Role: RealtorRole, ProfileURL: "https://0934608270.rieltor.ua"},
		}
		_ = SaveUsers()
	}

	// ВАЖЛИВО: Завжди завантажувати користувачів і парсити імена, навіть якщо файл щойно створено
	err := LoadUsers()
	if err != nil {
		log.Printf("Помилка завантаження користувачів: %v", err)
	}

	// Додаткове логування для перевірки завантаження користувачів
	fmt.Println("🔍 Завантажені користувачі:")
	for _, user := range users {
		fmt.Printf("👤 %+v\n", user)
	}
}

func ParseNameFromProfile(url string) (string, error) {
	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", fmt.Errorf("не вдалося створити запит: %v", err)
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) "+
		"AppleWebKit/537.36 (KHTML, like Gecko) Chrome/113.0.0.0 Safari/537.36")

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("не вдалося виконати запит: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return "", fmt.Errorf("не вдалося відкрити профіль: %v", resp.Status)
	}

	bodyBytes, _ := io.ReadAll(resp.Body)
	// Закоментовано після дебагу — при потребі можна розкоментувати:
	// fmt.Println("🧾 HTML сторінки профілю:\n", string(bodyBytes))

	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(bodyBytes))
	if err != nil {
		return "", fmt.Errorf("не вдалося парсити HTML: %v", err)
	}

	selection := doc.Find(".rieltor_card__name")
	if selection.Length() == 0 {
		return "", fmt.Errorf("не вдалося знайти елемент з класом rieltor_card__name")
	}

	name := strings.TrimSpace(selection.First().Text())
	if name == "" {
		return "", fmt.Errorf("не вдалося знайти ім’я на сторінці")
	}
	return name, nil
}

func PrintDebugInfo(users []User) string {
	var report strings.Builder
	report.WriteString("🧪 Поточний стан користувачів:\n")
	for _, u := range users {
		report.WriteString(fmt.Sprintf("- [%s] %d: %s (%s)\n", u.Role, u.TelegramID, u.Name, u.ProfileURL))
	}
	return report.String()
}
