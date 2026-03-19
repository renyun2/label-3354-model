package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/example/go-api-starter/internal/config"
)

var (
	once   sync.Once
	client *redis.Client
)

// Init 初始化Redis连接
func Init(cfg config.RedisConfig) (*redis.Client, error) {
	var err error
	once.Do(func() {
		client, err = newRedisClient(cfg)
	})
	return client, err
}

// Get 获取Redis客户端
func Get() *redis.Client {
	return client
}

func newRedisClient(cfg config.RedisConfig) (*redis.Client, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     cfg.Addr(),
		Password: cfg.Password,
		DB:       cfg.DB,
		PoolSize: cfg.PoolSize,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := rdb.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("连接Redis失败: %w", err)
	}

	return rdb, nil
}

// Set 设置键值（带过期时间）
func Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return client.Set(ctx, key, data, expiration).Err()
}

// Get 获取键值并反序列化
func GetJSON(ctx context.Context, key string, dest interface{}) error {
	data, err := client.Get(ctx, key).Bytes()
	if err != nil {
		return err
	}
	return json.Unmarshal(data, dest)
}

// Del 删除键
func Del(ctx context.Context, keys ...string) error {
	return client.Del(ctx, keys...).Err()
}

// Exists 检查键是否存在
func Exists(ctx context.Context, key string) (bool, error) {
	n, err := client.Exists(ctx, key).Result()
	return n > 0, err
}

// SetString 设置字符串键值
func SetString(ctx context.Context, key, value string, expiration time.Duration) error {
	return client.Set(ctx, key, value, expiration).Err()
}

// GetString 获取字符串键值
func GetString(ctx context.Context, key string) (string, error) {
	return client.Get(ctx, key).Result()
}

// IsNotFound 判断是否为键不存在错误
func IsNotFound(err error) bool {
	return err == redis.Nil
}
