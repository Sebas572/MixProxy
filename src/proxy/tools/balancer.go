package tools

import (
	"fmt"
	"mixproxy/src/proxy/config"
)

type ServerEntry struct {
	Turn     []int
	Petition int
}

type VpsProbability struct {
	Probability float64
	IP          string
}

var ServerSelected map[string]*ServerEntry = map[string]*ServerEntry{}

func SetupServerSelected(subdomain string, vpsProbability []VpsProbability) {
	probability := []float64{}
	for _, e := range vpsProbability {
		probability = append(probability, e.Probability)
	}

	turn := GenerateWRRSequence(probability)

	ServerSelected[subdomain] = &ServerEntry{
		Turn:     turn,
		Petition: 0,
	}
}

func GetTargetIPForSubdomain(subdomain string) (string, error) {
	if _, ok := ServerSelected[subdomain]; !ok {
		return "", fmt.Errorf("Subdomain not found")
	}

	i := ServerSelected[subdomain].Petition
	if i >= len(ServerSelected[subdomain].Turn) {
		ServerSelected[subdomain].Petition = 0
		i = 0
	}

	ServerSelected[subdomain].Petition++

	return config.Proxies[subdomain][ServerSelected[subdomain].Turn[i]], nil
}
