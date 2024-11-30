package api

import (
	"context"
	"log"
	"time"

	"github.com/redis/go-redis/v9"
)

const (
	redisConnString = "localhost:6379"
	defaultRedisDb  = 0
)

type RedisClient struct {
	redisClient *redis.Client
	ctx         context.Context
}

func NewRedisClient() *RedisClient {
	ctx := context.Background()
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
		ctx: ctx,
	}
}

// checks to see if the file data is in the Redis cache and, if so, returns it
//
// @param: cachedFileName - a string representing the cached file name,
//  which will be the project_name and file_name concatenated together.
//  this is done to ensure that the correct file from the correct project is returned.
// @return []byte - the file data, or nil if it is not in the cache
func (rc *RedisClient) getDataFromRedisCache(key string) []byte {
	d, err := rc.redisClient.Get(rc.ctx, key).Result()
	if err != nil {
		// an error here just means that the value is not cached, and that is ok
		log.Default().Printf("files.go::upload - redis error on get %s: %v", key, err)
	}

	if d != "" {
		log.Default().Printf("files.go::upload - found %s in redis cache with len %d", key, len([]byte(d)))
		return []byte(d)
	}

	return nil
}

func (rc *RedisClient) setDataInRedisCache(key string, data []byte) {
	err := redisClient.redisClient.Set(rc.ctx, key, data, time.Duration(24 * time.Hour)).Err()
	if err != nil {
		log.Default().Printf("files.go::upload - redis error on set %s: %v", key, err)
	}
}

func (rc *RedisClient) getFileMetaDataFromRedisCache(key string) map[string]string {
	var metadata map[string]string
	metadata, err := redisClient.redisClient.HGetAll(rc.ctx, key+"-metadata").Result()
	if err != nil || len(metadata) == 0 {
		log.Default().Printf("got error getting file metadata from redis cache. Err: %v", err)
	} else {
		ttl, err := redisClient.redisClient.TTL(rc.ctx, key).Result()
		if err != nil {
			log.Default().Printf("got error getting file metadata from redis cache. Err: %v", err)
		} else {
			metadata["ttl"] = ttl.String()
		}
	}

	return metadata
}

func (rc *RedisClient) setFileMetaDataInRedisCache(
	key, 
	filePath string,
	lastMtime,
	createdAt int64,
) {
	err := redisClient.redisClient.HMSet(rc.ctx, key+"-metadata", map[string]interface{}{
		"mtime": lastMtime,
		"createdAt": createdAt,
		"filePath": filePath,
		"ttl": time.Duration(24 * time.Hour).String(),
	}).Err()

	if err != nil {
		log.Default().Printf("error setting file metadata in Redis. Err: %v", err)
	}
}

func (rc *RedisClient) deleteFileAndMetadataFromRedisCache(fileName, metaDataName string) error {
	numKeysDeleted, err := redisClient.redisClient.Del(rc.ctx, fileName, metaDataName).Result()
	if err != nil || numKeysDeleted != 2 {
		log.Default().Printf("files.go::deleteFile - could not delete from redis cache. Err: %v", err)
		return err
	}

	return nil
}