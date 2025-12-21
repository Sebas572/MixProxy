package redis

import (
	"context"
	"encoding/json"
	"time"

	rd "github.com/redis/go-redis/v9"
)

var ctx = context.Background()
var rdb *rd.Client

type CachedResponse struct {
	Status  int               `json:"status"`
	Headers map[string]string `json:"headers"`
	Body    string            `json:"body"`
}

func init() {
	rdb = rd.NewClient(&rd.Options{
		Addr:     "redis:6379",
		Password: "mixproxy123",
		DB:       0,
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
