package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	rd "github.com/redis/go-redis/v9"
)

var ctx = context.Background()
var rdb *rd.Client      // DB 0 for other cache operations
var rdbStats *rd.Client // DB 1 for stats
var rdbLogs *rd.Client  // DB 2 for request logs
var rdbIPs *rd.Client   // DB 3 for IP stats

type CachedResponse struct {
	Status  int               `json:"status"`
	Headers map[string]string `json:"headers"`
	Body    string            `json:"body"`
}

type RequestLog struct {
	ID        string    `json:"id"`
	Method    string    `json:"method"`
	URL       string    `json:"url"`
	IP        string    `json:"ip"`
	Subdomain string    `json:"subdomain"`
	Timestamp time.Time `json:"timestamp"`
	Status    int       `json:"status"`
}

type IPStat struct {
	IP       string    `json:"ip"`
	Count    int       `json:"count"`
	LastSeen time.Time `json:"lastSeen"`
}

type Stats struct {
	TotalRequests     int `json:"totalRequests"`
	ActiveConnections int `json:"activeConnections"`
	UniqueIPs         int `json:"uniqueIPs"`
}

func init() {
	rdb = rd.NewClient(&rd.Options{
		Addr:     "redis:6379",
		Password: "mixproxy123",
		DB:       0,
	})
	rdbStats = rd.NewClient(&rd.Options{
		Addr:     "redis:6379",
		Password: "mixproxy123",
		DB:       1,
	})
	rdbLogs = rd.NewClient(&rd.Options{
		Addr:     "redis:6379",
		Password: "mixproxy123",
		DB:       2,
	})
	rdbIPs = rd.NewClient(&rd.Options{
		Addr:     "redis:6379",
		Password: "mixproxy123",
		DB:       3,
	})
}

func SetCachePage(url, content string) error {
	return rdb.Set(ctx, url, content, -1).Err()
}

func GetCachePage(url string) (string, error) {
	content, err := rdb.Get(ctx, url).Result()
	if err != nil {
		return "", err
	}
	return content, nil
}

func SetBlockIP(ip string) error {
	return rdb.Set(ctx, ip, true, 30*time.Minute).Err()
}

func IsTheIPBlocked(ip string) (bool, error) {
	result, err := rdb.Get(ctx, ip).Bool()
	if err != nil {
		return false, err
	}
	return result, nil
}

func SetCachedResponse(key string, resp CachedResponse, ttl time.Duration) error {
	data, err := json.Marshal(resp)
	if err != nil {
		return err
	}
	return rdb.Set(ctx, key, data, ttl).Err()
}

func GetCachedResponse(key string) (CachedResponse, bool, error) {
	data, err := rdb.Get(ctx, key).Result()
	if err == rd.Nil {
		return CachedResponse{}, false, nil
	}
	if err != nil {
		return CachedResponse{}, false, err
	}
	var resp CachedResponse
	err = json.Unmarshal([]byte(data), &resp)
	if err != nil {
		return CachedResponse{}, false, err
	}
	return resp, true, nil
}

func SetAllowSubdomainToUseCache(subdomain string, value bool) error {
	return rdb.Set(ctx, subdomain, value, -1).Err()
}

func DoesTheSubdomainAllowCache(subdomain string) bool {
	isAllow, err := rdb.Get(ctx, subdomain).Bool()
	if err != nil {
		return false
	}

	return isAllow
}

func SetCachePaths(subdomain string, paths []string) error {
	data, err := json.Marshal(paths)
	if err != nil {
		return err
	}
	return rdb.Set(ctx, "cache_paths:"+subdomain, data, -1).Err()
}

func GetCachePaths(subdomain string) ([]string, error) {
	data, err := rdb.Get(ctx, "cache_paths:"+subdomain).Result()
	if err == rd.Nil {
		return []string{}, nil
	}
	if err != nil {
		return nil, err
	}
	var paths []string
	err = json.Unmarshal([]byte(data), &paths)
	if err != nil {
		return nil, err
	}
	return paths, nil
}

func SetStats(stats Stats) error {
	data, err := json.Marshal(stats)
	if err != nil {
		return err
	}
	return rdbStats.Set(ctx, "stats", data, -1).Err()
}

func GetStats() (Stats, error) {
	data, err := rdbStats.Get(ctx, "stats").Result()
	if err == rd.Nil {
		return Stats{}, nil
	}
	if err != nil {
		return Stats{}, err
	}
	var stats Stats
	err = json.Unmarshal([]byte(data), &stats)
	return stats, err
}

func SetRequestLogs(logs []RequestLog) error {
	data, err := json.Marshal(logs)
	if err != nil {
		return err
	}
	return rdbLogs.Set(ctx, "request_logs", data, -1).Err()
}

func GetRequestLogs() ([]RequestLog, error) {
	data, err := rdbLogs.Get(ctx, "request_logs").Result()
	if err == rd.Nil {
		return []RequestLog{}, nil
	}
	if err != nil {
		return nil, err
	}
	var logs []RequestLog
	err = json.Unmarshal([]byte(data), &logs)
	return logs, err
}

func SetIPStats(stats []IPStat) error {
	data, err := json.Marshal(stats)
	if err != nil {
		return err
	}
	return rdbIPs.Set(ctx, "ip_stats", data, -1).Err()
}

func GetIPStats() ([]IPStat, error) {
	data, err := rdbIPs.Get(ctx, "ip_stats").Result()
	if err == rd.Nil {
		return []IPStat{}, nil
	}
	if err != nil {
		return nil, err
	}
	var ipStats []IPStat
	err = json.Unmarshal([]byte(data), &ipStats)
	return ipStats, err
}

func AddRequestLog(method, url, ip, subdomain string, status int) error {
	// Get current logs
	currentLogs, err := GetRequestLogs()
	if err != nil {
		return err
	}

	// Create new log
	newLog := RequestLog{
		ID:        fmt.Sprintf("%d", time.Now().UnixNano()),
		Method:    method,
		URL:       url,
		IP:        ip,
		Subdomain: subdomain,
		Timestamp: time.Now(),
		Status:    status,
	}

	// Append
	currentLogs = append(currentLogs, newLog)

	// Keep only last 100
	if len(currentLogs) > 100 {
		currentLogs = currentLogs[1:]
	}

	// Set logs back
	logsData, err := json.Marshal(currentLogs)
	if err != nil {
		return err
	}
	if err := rdbLogs.Set(ctx, "request_logs", logsData, -1).Err(); err != nil {
		return err
	}

	// Update stats
	currentStats, err := GetStats()
	if err != nil {
		return err
	}
	currentStats.TotalRequests++

	// Update IP stats
	currentIPStats, err := GetIPStats()
	if err != nil {
		return err
	}

	ipKey := strings.Split(ip, ":")[0]
	found := false
	for i, stat := range currentIPStats {
		if stat.IP == ipKey {
			currentIPStats[i].Count++
			currentIPStats[i].LastSeen = time.Now()
			found = true
			break
		}
	}
	if !found {
		newIPStat := IPStat{
			IP:       ipKey,
			Count:    1,
			LastSeen: time.Now(),
		}
		currentIPStats = append(currentIPStats, newIPStat)
		currentStats.UniqueIPs = len(currentIPStats)
	}

	// Set stats back
	statsData, err := json.Marshal(currentStats)
	if err != nil {
		return err
	}
	if err := rdbStats.Set(ctx, "stats", statsData, -1).Err(); err != nil {
		return err
	}

	// Set IP stats back
	ipStatsData, err := json.Marshal(currentIPStats)
	if err != nil {
		return err
	}
	return rdbIPs.Set(ctx, "ip_stats", ipStatsData, -1).Err()
}

func Clean() {
	err := rdb.FlushDB(ctx).Err()
	if err != nil {
		panic(err)
	}
}
