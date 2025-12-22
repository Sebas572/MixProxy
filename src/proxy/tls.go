package proxy

func getTlsConfig() (string, string) {
	return "./certs/wildcard.crt", "./certs/wildcard.key"
}
