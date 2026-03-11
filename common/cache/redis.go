package cache

import (
	"context"
	"time"

	"github.com/luckysxx/user-platform/common/config"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

// Cache 接口定义缓存操作
type Cache interface {
	Get(ctx context.Context, key string) (string, error)
	Set(ctx context.Context, key string, value string, expiration time.Duration) error
	Del(ctx context.Context, keys ...string) error
}

// redisCache 实现 Cache 接口
type redisCache struct {
	client *redis.Client
}

// NewCache 创建 Cache 实例
func NewCache(client *redis.Client) Cache {
	return &redisCache{client: client}
}

func (r *redisCache) Get(ctx context.Context, key string) (string, error) {
	return r.client.Get(ctx, key).Result()
}

func (r *redisCache) Set(ctx context.Context, key string, value string, expiration time.Duration) error {
	return r.client.Set(ctx, key, value, expiration).Err()
}

func (r *redisCache) Del(ctx context.Context, keys ...string) error {
	return r.client.Del(ctx, keys...).Err()
}

// InitRedis 初始化 Redis 客户端
func InitRedis(cfg config.RedisConfig, log *zap.Logger) *redis.Client {
	client := redis.NewClient(&redis.Options{
		Addr:     cfg.Addr,
		Password: cfg.Password,
		DB:       cfg.DB,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		log.Fatal("无法连接到 Redis", zap.Error(err))
		return nil
	}

	log.Info("成功连接到 Redis")
	return client
}
