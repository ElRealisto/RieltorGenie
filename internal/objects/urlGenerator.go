package objects

import (
	"bytes"
	"fmt"
	"log"
	"os/exec"
)

// GenerateStreetURL запускає Playwright-скрипт для отримання URL
func GenerateStreetURL() {
	cmd := exec.Command("node", "scripts/generateStreetURL.js")

	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		log.Fatalf("Помилка: %s\n%s", err, stderr.String())
	}

	finalURL := out.String()
	fmt.Printf("✅ Отриманий URL: %s\n", finalURL)
}
