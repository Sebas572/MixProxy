// Codigo unicamente para pruebas locales de certificados

package certificate

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"os"
	"strings"
	"time"
)

func Create(DNSnames []string) {
	os.MkdirAll("certs", 0755)

	privKey, _ := rsa.GenerateKey(rand.Reader, 2048)

	template := x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			Organization: []string{"Dev Local"},
			CommonName:   "*.developer.space",
		},
		DNSNames:    DNSnames, //[]string{"developer.space", "test.developer.space", "app.developer.space"},
		NotBefore:   time.Now(),
		NotAfter:    time.Now().Add(365 * 24 * time.Hour),
		KeyUsage:    x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
	}

	derBytes, _ := x509.CreateCertificate(rand.Reader, &template, &template, &privKey.PublicKey, privKey)

	certOut, _ := os.Create("certs/wildcard.developer.space.crt")
	pem.Encode(certOut, &pem.Block{Type: "CERTIFICATE", Bytes: derBytes})
	certOut.Close()

	keyOut, _ := os.Create("certs/wildcard.developer.space.key")
	pem.Encode(keyOut, &pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(privKey)})
	keyOut.Close()

	fmt.Println("✅ Generated wildcard certificate")
	fmt.Println(strings.Join(DNSnames, "\n✅"))
}
