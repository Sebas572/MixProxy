package redis

import (
	"encoding/json"
	"strings"
	"time"
)

type Reason struct {
	Content string
	Time    time.Time
	Date    string
}

func EnabledBlacklistForSubdomain(subdomain string) error {
	return rdbBlacklist.Set(ctx, subdomain, true, -1).Err()
}

func DisabledBlacklistForSubdomain(subdomain string) error {
	return rdbBlacklist.Set(ctx, subdomain, false, -1).Err()
}

func IsEnabledBlacklistForSubdomain(subdomain string) (bool, error) {
	isEnabled, err := rdbBlacklist.Get(ctx, subdomain).Bool()
	if err != nil {
		return false, err
	}

	return isEnabled, nil
}

func SetIPForBlacklist(subdomain, ip string, reason Reason, duration time.Duration) error {
	data, err := json.Marshal(reason)
	if err != nil {
		return err
	}
	return rdbBlacklist.Set(ctx, "["+subdomain+"]"+ip, string(data), duration).Err()
}

func GetIPForBlacklist(subdomain, ip string) (Reason, error) {
	content, err := rdbBlacklist.Get(ctx, "["+subdomain+"]"+ip).Result()
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

func RemoveIPFromBlacklist(subdomain, ip string) error {
	return rdbBlacklist.Del(ctx, "["+subdomain+"]"+ip).Err()
}

func GetAllIPsForBlacklist(subdomain string) (map[string]Reason, error) {
	keys, err := rdbBlacklist.Keys(ctx, "[[]"+subdomain+"]*").Result()
	if err != nil {
		return nil, err
	}
	result := make(map[string]Reason)
	for _, key := range keys {
		ip := strings.TrimPrefix(key, "["+subdomain+"]")
		reason, err := GetIPForBlacklist(subdomain, ip)
		if err != nil {
			continue
		}
		result[ip] = reason
	}
	return result, nil
}

func GetAllEnabledBlacklistSubdomains() ([]string, error) {
	keys, err := rdbBlacklist.Keys(ctx, "*").Result()
	if err != nil {
		return nil, err
	}
	var subdomains []string
	for _, key := range keys {
		if !strings.HasPrefix(key, "[") && !strings.HasPrefix(key, "global:") { // not an IP key or global
			enabled, err := rdbBlacklist.Get(ctx, key).Bool()
			if err == nil && enabled {
				subdomains = append(subdomains, key)
			}
		}
	}
	return subdomains, nil
}

func SetIPForGlobalBlacklist(ip string, reason Reason, duration time.Duration) error {
	data, err := json.Marshal(reason)
	if err != nil {
		return err
	}
	return rdbBlacklist.Set(ctx, "global:"+ip, string(data), duration).Err()
}

func GetIPForGlobalBlacklist(ip string) (Reason, error) {
	content, err := rdbBlacklist.Get(ctx, "global:"+ip).Result()
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

func RemoveIPFromGlobalBlacklist(ip string) error {
	return rdbBlacklist.Del(ctx, "global:"+ip).Err()
}

func GetAllIPsForGlobalBlacklist() (map[string]Reason, error) {
	keys, err := rdbBlacklist.Keys(ctx, "global:*").Result()
	if err != nil {
		return nil, err
	}
	result := make(map[string]Reason)
	for _, key := range keys {
		ip := strings.TrimPrefix(key, "global:")
		reason, err := GetIPForGlobalBlacklist(ip)
		if err != nil {
			continue
		}
		result[ip] = reason
	}
	return result, nil
}
