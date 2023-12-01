package cache_redis

import (
	"time"
	"context"
	"github.com/rs/zerolog/log"

	redis "github.com/redis/go-redis/v9"
	"github.com/aws/aws-xray-sdk-go/xray"
)

var childLogger = log.With().Str("repository/cache", "Redis").Logger()

type CacheService struct {
	cache *redis.ClusterClient
	cache_s *redis.Client
}

func NewCache(ctx context.Context, options *redis.Options) *CacheService {
	childLogger.Debug().Msg("NewCache")
	childLogger.Debug().Interface("option.Addr:", options.Addr).Msg("")

	redisClient := redis.NewClient(options)
	return &CacheService{
		cache_s: redisClient,
	}
}

func NewClusterCache(ctx context.Context, options *redis.ClusterOptions) *CacheService {
	childLogger.Debug().Msg("NewClusterCache")
	childLogger.Debug().Interface("option.Addrs: ", options.Addrs).Msg("")

	redisClient := redis.NewClusterClient(options)
	return &CacheService{
		cache: redisClient,
	}
}

func (s *CacheService) Sum(ctx context.Context, key string, value interface{}) (error) {
	childLogger.Debug().Msg("Sum")

	_, root := xray.BeginSubsegment(ctx, "REDIS.HIncrByFloat-Balance-Charges")
	defer func() {
		root.Close(nil)
	}()

	_, err := s.cache.HIncrByFloat(ctx, "account:" + key, "amount", value.(float64)).Result()
	if err != nil {
		return err
	}

	s.cache.PExpire(ctx, "account:" + key, time.Minute * 1).Result()

	return nil
}

func (s *CacheService) Get(ctx context.Context, key string) (interface{}, error) {
	childLogger.Debug().Msg("Get")

	_, root := xray.BeginSubsegment(ctx, "REDIS.HGet-Balance-Charges")
	defer func() {
		root.Close(nil)
	}()

	res, err := s.cache.HGet(ctx, "account:"+ key, "amount").Result()
	if err != nil {
		return nil, err
	}

	childLogger.Debug().Interface("+++++ RES : ",res).Msg("Get")

	return res, nil
}

func (s *CacheService) Put(ctx context.Context, key string, value interface{}) error {
	childLogger.Debug().Msg("Put")
	
	_, root := xray.BeginSubsegment(ctx, "REDIS.Set-Balance-Charges")
	defer func() {
		root.Close(nil)
	}()

	status := s.cache.Set(ctx, "account:"+ key, value, time.Minute * 10)
	return status.Err()
}

func (s *CacheService) Ping(ctx context.Context) (string, error) {
	childLogger.Debug().Msg("Ping")

	status, err := s.cache.Ping(ctx).Result()
	if err != nil {
		return "", err
	}
	return status, nil
}
