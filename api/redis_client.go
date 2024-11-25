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