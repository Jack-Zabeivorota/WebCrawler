package cache

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/redis/go-redis/v9"

	"main/logger"
	"main/models"
)

type RedisCache struct {
	client *redis.Client
	log    *logger.Logger
}

func NewRedisCache(host string) *RedisCache {
	if host == "" {
		host = "localhost:6379"
	}

	return &RedisCache{
		client: redis.NewClient(&redis.Options{
			Addr: host,
		}),
		log: logger.Instance(),
	}
}

func (rc *RedisCache) SetRequestData(requestID int64, data *models.RequestData) error {
	ctx := context.Background()

	value, err := json.Marshal(data)

	if err != nil {
		rc.log.Error("Request %d (SetRequestData) -> Request data parsing error: %v", requestID, err)
		return err
	}

	err = rc.client.HSet(ctx, "requests_data", requestID, value).Err()

	if err != nil {
		rc.log.Error("Request %d -> Setting to request data to cache error: %v", requestID, err)
	}

	return err
}

func (rc *RedisCache) SetURLToAllURLs(requestID int64, url string) error {
	ctx := context.Background()
	id := fmt.Sprintf("all_urls:%d", requestID)

	err := rc.client.SAdd(ctx, id, url).Err()

	if err != nil {
		rc.log.Error("Request %d -> Setting URL to all URLs error: %v", requestID, err)
	}

	return err
}

func (rc *RedisCache) ClearRequest(requestID int64) error {
	ctx := context.Background()

	allUrlKey := fmt.Sprintf("all_urls:%d", requestID)
	completedUrlKey := fmt.Sprintf("completed_urls:%d", requestID)
	id := fmt.Sprintf("%d", requestID)

	err := rc.client.Watch(ctx, func(tx *redis.Tx) error {
		_, err := tx.TxPipelined(ctx, func(pipe redis.Pipeliner) error {

			err := pipe.Del(ctx, completedUrlKey, allUrlKey).Err()

			if err != nil {
				return err
			}

			return pipe.HDel(ctx, "requests_data", id).Err()
		})
		return err
	}, allUrlKey, completedUrlKey, "requests_data")

	if err != nil {
		rc.log.Error("Request %d -> Clearing request in cache error: %v", requestID, err)
	}

	return err
}
