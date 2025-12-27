package main

import (
	"encoding/base64"
	"fmt"
	"io"
	"mixproxy/src/proxy/config"
	"net/http"
	"strings"
	"time"
)

const (
	Reset  = "\033[0m"
	Red    = "\033[31m"
	Green  = "\033[32m"
	Yellow = "\033[33m"
)

func printStatus(success bool) {
	if success {
		fmt.Printf("%s✓%s\n", Green, Reset)
	} else {
		fmt.Printf("%s✗%s\n", Red, Reset)
	}
}

func makeRequest(method, url string, expectedStatus int, description string) bool {
	fmt.Printf("Testing %s... ", description)

	client := &http.Client{
		Timeout: 10 * time.Second,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		printStatus(false)
		return false
	}

	if strings.Contains(url, "admin") {
		encoded := base64.StdEncoding.EncodeToString([]byte(config.AdminUsername + ":" + config.AdminPassword))
		req.Header.Set("Authorization", "Basic "+encoded)
	}

	resp, err := client.Do(req)
	if err != nil {
		printStatus(false)
		return false
	}
	defer resp.Body.Close()

	success := resp.StatusCode == expectedStatus
	printStatus(success)

	if !success {
		body, _ := io.ReadAll(resp.Body)
		fmt.Printf("      Expected status %d, got %d. Response: %s\n", expectedStatus, resp.StatusCode, string(body))
	}

	return success
}

func main() {
	fmt.Println("=== Starting MixProxy Integration Tests ===")
	fmt.Println()

	time.Sleep(2 * time.Second) // Wait a bit

	cfg, err := config.ReadConfig()

	if err != nil {

		fmt.Println("Error reading config:", err)

		return

	}

	config.AdminUsername = cfg.AdminUsername

	config.AdminPassword = cfg.AdminPassword

	client := &http.Client{
		Timeout: 15 * time.Second,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
	var resp *http.Response

	// 1. Verificar servidor corriendo (HTTP redirect to HTTPS)
	fmt.Printf("Testing Server running... ")
	req, err := http.NewRequest("GET", "http://"+cfg.Hostname+"", nil)
	serverRunning := false
	if err == nil {
		resp, err = client.Do(req)
		if err == nil {
			defer resp.Body.Close()
			serverRunning = resp.StatusCode == 200 || (resp.StatusCode == 301 || resp.StatusCode == 302) && strings.HasPrefix(resp.Header.Get("Location"), "https://")
		}
	}
	printStatus(serverRunning)

	// 2. Panel de administrador detectado
	adminPanelRunning := makeRequest("GET", "https://"+cfg.SubdomainAdminPanel+"."+cfg.Hostname+"", 200, "Admin panel detected")

	// 3. Verificar el admin-api
	adminAPIRunning := makeRequest("GET", "https://admin-api."+cfg.Hostname+"/api/config", 200, "Verify admin-api")

	// 4. Verificar páginas según configuración de subdominios y root
	fmt.Printf("Testing Verify if pages according to subdomain and root configuration are running... ")
	subdomainsRunning := false

	// Test root domain (should redirect or serve)
	req, err = http.NewRequest("GET", "https://"+cfg.Hostname+"", nil)
	if err == nil {
		resp, err = client.Do(req)
		if err == nil {
			defer resp.Body.Close()
			if resp.StatusCode == 200 || resp.StatusCode == 301 || resp.StatusCode == 302 {
				subdomainsRunning = true
			}
		}
	}

	// Test admin subdomain
	req, err = http.NewRequest("GET", "https://"+cfg.SubdomainAdminPanel+"."+cfg.Hostname+"", nil)
	if err == nil {
		encoded := base64.StdEncoding.EncodeToString([]byte(config.AdminUsername + ":" + config.AdminPassword))
		req.Header.Set("Authorization", "Basic "+encoded)
		resp, err = client.Do(req)
		if err == nil {
			defer resp.Body.Close()
			if resp.StatusCode == 200 || resp.StatusCode == 301 || resp.StatusCode == 302 {
				subdomainsRunning = true
			}
		}
	}

	printStatus(subdomainsRunning)

	// Additional checks
	configRunning := makeRequest("GET", "https://admin-api."+cfg.Hostname+"/api/config", 200, "Verifying config API")

	logsRunning := makeRequest("GET", "https://admin-api."+cfg.Hostname+"/api/logs/list", 200, "Verifying logs API")

	statsRunning := makeRequest("GET", "https://admin-api."+cfg.Hostname+"/api/stats", 200, "Verifying stats API")

	requestsRunning := makeRequest("GET", "https://admin-api."+cfg.Hostname+"/api/requests", 200, "Verifying requests API")

	ipsRunning := makeRequest("GET", "https://admin-api."+cfg.Hostname+"/api/ips", 200, "Verifying ips API")

	reloadRunning := makeRequest("POST", "https://admin-api."+cfg.Hostname+"/api/reload", 200, "Verifying reload API")

	// Summary
	fmt.Println()
	fmt.Println("=== Test Summary ===")
	allTests := []bool{serverRunning, adminPanelRunning, adminAPIRunning, subdomainsRunning, configRunning, logsRunning, statsRunning, requestsRunning, ipsRunning, reloadRunning}

	passed := 0
	for _, test := range allTests {
		if test {
			passed++
		}
	}

	fmt.Printf("Tests passed: %d/%d\n", passed, len(allTests))

	if passed == len(allTests) {
		fmt.Println("🎉 All tests passed successfully!")
	} else {
		fmt.Println("❌ Some tests failed. Check configuration and services.")
	}
}
