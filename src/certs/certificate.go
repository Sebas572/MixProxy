// Codigo unicamente para pruebas locales de certificados

package certificate

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

func Create(DNSnames []string) {
	os.MkdirAll("certs", 0755)

	// Install mkcert CA if not already installed
	installCmd := exec.Command("mkcert", "-install")
	installErr := installCmd.Run()
	if installErr != nil {
		fmt.Printf("Error installing mkcert CA: %v\n", installErr)
		fmt.Println("Please install mkcert: https://github.com/FiloSottile/mkcert")
		return
	}

	// Use mkcert to generate trusted certificates
	args := []string{"-cert-file", "certs/wildcard.crt", "-key-file", "certs/wildcard.key"}
	args = append(args, DNSnames...)

	cmd := exec.Command("mkcert", args...)
	err := cmd.Run()
	if err != nil {
		fmt.Printf("Error running mkcert: %v\n", err)
		fmt.Println("Please install mkcert: https://github.com/FiloSottile/mkcert")
		return
	}

	fmt.Println("✅ Generated trusted certificates with mkcert")
	fmt.Println("✅ " + strings.Join(DNSnames, "\n✅ "))
}
