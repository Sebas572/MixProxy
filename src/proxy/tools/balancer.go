package tools

import (
	"mixproxy/src/proxy/config"
	"net/http/httputil"
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

func GetTargetIPForSubdomain(subdomain string) (*httputil.ReverseProxy, error) {
	var target *httputil.ReverseProxy

	i := ServerSelected[subdomain].Petition
	if i >= len(ServerSelected[subdomain].Turn) {
		ServerSelected[subdomain].Petition = 0
		i = 0
	}

	target = config.Proxies[subdomain][i]
	ServerSelected[subdomain].Petition++

	return target, nil
}
