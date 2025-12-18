package config

import (
	"fmt"
	"os"
)

var TEXT_FILE_CONFIG string = "{\n\t\"hostname\": \"%s\",\n\t\"subdomain_admin_panel\": \"%s\",\n\t\"on_https\": %s,\n\t\"mode_developer\": %s,\n\t\"load_balancer\": [],\n\t\"root_load_balancer\": {}\n}"

func input(ask, defaultValue string) string {
	fmt.Printf("%s (default: %s): ", ask, defaultValue)

	var input string
	fmt.Scanln(&input)
	if input == "" {
		return defaultValue
	}
	return input
}

func CreateConfig() {
	os.Mkdir(".config", 0755)

	hostname := input("Enter hostname", "developer.space")
	subdomain_admin_panel := input("Enter subdomain admin panel", "admin")
	on_https := input("On HTTPS? (true/false)", "true")
	mode_developer := input("Mode developer? (true/false)", "true")

	text := []byte(fmt.Sprintf(TEXT_FILE_CONFIG, hostname, subdomain_admin_panel, on_https, mode_developer))

	if err := os.WriteFile(CONFIG_PATH, text, 0644); err != nil {
		fmt.Println("Error to create proxy.config.json")
	}
}
