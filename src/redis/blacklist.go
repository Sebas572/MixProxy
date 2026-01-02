package redis

import (
	"encoding/json"
	"fmt"
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

func ChangeSubdomainBlacklist(oldSubdomain, newSubdomain string) {
	fmt.Printf("%s -> %s\n", oldSubdomain, newSubdomain)
	if oldSubdomain == newSubdomain {
		fmt.Println("Subdomains are the same")
		return
	}

	pattern := fmt.Sprintf("[[]%s]*", oldSubdomain)
	var cursor uint64
	var changed int

	for {
		// Buscar keys
		keys, nextCursor, _ := rdbBlacklist.Scan(ctx, cursor, pattern, 100).Result()
		fmt.Println(keys)

		// Procesar keys
		for _, oldKey := range keys {
			parts := strings.SplitN(oldKey, "]", 2)
			if len(parts) != 2 {
				continue
			}

			newKey := fmt.Sprintf("[%s]%s", newSubdomain, parts[1])
			if oldKey == newKey {
				continue
			}

			// Intentar renombrar, si falla ignorar
			if err := rdbBlacklist.Rename(ctx, oldKey, newKey).Err(); err == nil {
				changed++
			}
		}

		if nextCursor == 0 {
			break
		}
		cursor = nextCursor
	}

	if err := rdbBlacklist.Rename(ctx, oldSubdomain, newSubdomain).Err(); err == nil {
		changed++
	}

	fmt.Printf("Changed %d keys from [%s] to [%s]\n", changed, oldSubdomain, newSubdomain)
}

func RemoveAllForBlacklistSubdomain(subdomain string) error {
	// Get all IPs
	ips, err := GetAllIPsForBlacklist(subdomain)
	if err != nil {
		return err
	}
	// Delete each IP
	for ip := range ips {
		RemoveIPFromBlacklist(subdomain, ip)
	}
	// Delete enabled
	return DisabledBlacklistForSubdomain(subdomain)
}
