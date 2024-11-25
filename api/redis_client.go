package api

import (
	"context"
	"log"

	"github.com/redis/go-redis/v9"
)

const (
	redisConnString = "localhost:6379"
	defaultRedisDb  = 0
)

var (
	ctx = context.Background()
)

type RedisClient struct {
	redisClient *redis.Client
}

func NewRedisClient() *RedisClient {
	redisClient := redis.NewClient(&redis.Options{
		Addr: redisConnString,
		DB: defaultRedisDb,
		Password: "",
		PoolSize: 20,
	})

	_, err := redisClient.Ping(ctx).Result()
	if err != nil {
		log.Fatalf("redis_client.go::New - could not connect to redis: %v", err)
	}

	log.Default().Println("redis_client.go::Connected to Redis!")
	return &RedisClient{
		redisClient: redisClient,
	}
}

// checks to see if the file data is in the Redis cache and, if so, returns it
//
// @param: cachedFileName - a string representing the cached file name,
//  which will be the project_name and file_name concatenated together.
//  this is done to ensure that the correct file from the correct project is returned.
// @return []byte - the file data, or nil if it is not in the cache
func (rc *RedisClient) getFileFromRedisCache(cachedFileName string) []byte {
	d, err := rc.redisClient.Get(ctx, cachedFileName).Result()
	if err != nil {
		log.Default().Printf("files.go::upload - redis error on get %s: %v", cachedFileName, err)
	}

	if d != "" {
		log.Default().Printf("files.go::upload - found %s in redis cache with len %d", cachedFileName, len([]byte(d)))
		return []byte(d)
	}

	return nil
}