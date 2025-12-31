package redis

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

func EnabledWhitelistForSubdomain(subdomain string) error {
	return rdbWhitelist.Set(ctx, subdomain, true, -1).Err()
}

func DisabledWhitelistForSubdomain(subdomain string) error {
	return rdbWhitelist.Set(ctx, subdomain, false, -1).Err()
}

func IsEnabledWhitelistForSubdomain(subdomain string) (bool, error) {
	isEnabled, err := rdbWhitelist.Get(ctx, subdomain).Bool()
	if err != nil {
		return false, err
	}

	return isEnabled, nil
}

func SetIPForWhitelist(subdomain, ip string, reason Reason, duration time.Duration) error {
	data, err := json.Marshal(reason)
	if err != nil {
		return err
	}
	return rdbWhitelist.Set(ctx, "["+subdomain+"]"+ip, string(data), duration).Err()
}

func GetIPForWhitelist(subdomain, ip string) (Reason, error) {
	fmt.Println("[" + subdomain + "]" + ip)
	content, err := rdbWhitelist.Get(ctx, "["+subdomain+"]"+ip).Result()

	if err != nil {
		return Reason{}, err
	}
	var reason Reason
	err = json.Unmarshal([]byte(content), &reason)
	if err != nil {
		return Reason{}, err
	}
	return reason, nil
}

func RemoveIPFromWhitelist(subdomain, ip string) error {
	return rdbWhitelist.Del(ctx, "["+subdomain+"]"+ip).Err()
}

func GetAllIPsForWhitelist(subdomain string) (map[string]Reason, error) {
	keys, err := rdbWhitelist.Keys(ctx, "[[]"+subdomain+"]*").Result()
	if err != nil {
		return nil, err
	}
	result := make(map[string]Reason)
	for _, key := range keys {
		ip := strings.TrimPrefix(key, "["+subdomain+"]")
		reason, err := GetIPForWhitelist(subdomain, ip)
		if err != nil {
			continue
		}
		result[ip] = reason
	}
	return result, nil
}

func GetAllEnabledWhitelistSubdomains() ([]string, error) {
	keys, err := rdbWhitelist.Keys(ctx, "*").Result()
	if err != nil {
		return nil, err
	}
	var subdomains []string
	for _, key := range keys {
		if !strings.HasPrefix(key, "[") { // not an IP key
			enabled, err := rdbWhitelist.Get(ctx, key).Bool()
			if err == nil && enabled {
				subdomains = append(subdomains, key)
			}
		}
	}
	return subdomains, nil
}
