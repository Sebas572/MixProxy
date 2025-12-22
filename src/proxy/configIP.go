package proxy

import (
	"log"
	"mixproxy/src/proxy/config"
	"mixproxy/src/proxy/tools"
	"mixproxy/src/redis"
	"os"
)

type LoadBalancer struct {
	URL      string
	Capacity float64
}

func reloadConfig() {
	config.Proxies = make(map[string][]string)
	tools.ServerSelected = make(map[string]*tools.ServerEntry)

	cfg, _ := config.ReadConfig()
	redis.Clean()

	if err := config.ValidateConfig(cfg); err != nil {
		log.Println("❌ Error de validación:", err)
		os.Exit(0)
	}

	loadBalancer := map[string]*[]LoadBalancer{}

	if cfg.ModeDeveloper {
		log.Println("Configuring certificates in development mode")
	}

	for _, e := range cfg.LoadBalancer {
		vps := []LoadBalancer{}
		subdomain := e.Subdomain
		redis.SetAllowSubdomainToUseCache(subdomain, e.CacheEnabled)
		redis.SetCachePaths(subdomain, e.CachePaths)

		for _, v := range e.VPS {
			vps = append(vps, LoadBalancer{
				URL:      v.IP,
				Capacity: v.Capacity,
			})
		}

		loadBalancer[subdomain] = &vps
		probability := []tools.VpsProbability{}
		for _, v := range e.VPS {
			probability = append(probability, tools.VpsProbability{
				Probability: v.Capacity,
				IP:          v.IP,
			})
		}
		tools.SetupServerSelected(subdomain, probability)
	}

	if cfg.RootLoadBalancer != nil && config.AllValuesNonEmpty(cfg.RootLoadBalancer) {
		vps := []LoadBalancer{}
		subdomain := ""
		redis.SetAllowSubdomainToUseCache(subdomain, cfg.RootLoadBalancer.CacheEnabled)
		redis.SetCachePaths(subdomain, cfg.RootLoadBalancer.CachePaths)

		for _, v := range cfg.RootLoadBalancer.VPS {
			vps = append(vps, LoadBalancer{
				URL:      v.IP,
				Capacity: v.Capacity,
			})
		}

		loadBalancer[subdomain] = &vps
		probability := []tools.VpsProbability{}
		for _, v := range cfg.RootLoadBalancer.VPS {
			probability = append(probability, tools.VpsProbability{
				Probability: v.Capacity,
				IP:          v.IP,
			})
		}
		tools.SetupServerSelected(subdomain, probability)
	}

	if len(loadBalancer) != 0 {
		for subdomain, targets := range loadBalancer {
			for _, target := range *targets {
				config.Proxies[subdomain] = append(config.Proxies[subdomain], target.URL)
			}
		}
	}
}
