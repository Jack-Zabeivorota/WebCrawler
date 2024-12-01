package cache

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"

	"main/logger"
	"main/tools"
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

func (rc *RedisCache) AllURLsIsCompleted(requestID int64) (bool, error) {
	ctx := context.Background()

	allUrlKey := fmt.Sprintf("all_urls:%d", requestID)
	completedUrlKey := fmt.Sprintf("completed_urls:%d", requestID)

	var allURLCount, completedUrlCount int64
	var err error

	tools.RetryCycle(func() error {
		allURLCount, err = rc.client.SCard(ctx, allUrlKey).Result()
		return err
	}, "Redis fail try", false)

	if err != nil {
		rc.log.Error("Request %d -> All URLs is completed checking error: %v", requestID, err)
		return false, err
	}

	tools.RetryCycle(func() error {
		completedUrlCount, err = rc.client.HLen(ctx, completedUrlKey).Result()
		return err
	}, "Redis fail try", false)

	if err != nil {
		rc.log.Error("Request %d -> All URLs is completed checking error: %v", requestID, err)
		return false, err
	}

	return allURLCount == completedUrlCount, nil
}

func (rc *RedisCache) GetURLsResult(requestID int64) (map[string]string, error) {
	ctx := context.Background()
	id := fmt.Sprintf("completed_urls:%d", requestID)

	var results map[string]string
	var err error

	tools.RetryCycle(func() error {
		results, err = rc.client.HGetAll(ctx, id).Result()
		return err
	}, "Redis fail try", false)

	if err != nil {
		rc.log.Error("Request %d -> Getting URLs results error: %v", requestID, err)
	}

	return results, err
}

func (rc *RedisCache) ClearRequest(requestID int64) error {
	ctx := context.Background()

	allUrlKey := fmt.Sprintf("all_urls:%d", requestID)
	completedUrlKey := fmt.Sprintf("completed_urls:%d", requestID)
	id := fmt.Sprintf("%d", requestID)

	err := tools.RetryCycle(func() error {
		return rc.client.Watch(ctx, func(tx *redis.Tx) error {
			_, err := tx.TxPipelined(ctx, func(pipe redis.Pipeliner) error {

				err := pipe.Del(ctx, completedUrlKey, allUrlKey).Err()

				if err != nil {
					return err
				}

				return pipe.HDel(ctx, "requests_data", id).Err()

			})
			return err
		}, allUrlKey, completedUrlKey, "requests_data")
	}, "Redis fail try", false)

	if err != nil {
		rc.log.Error("Request %d -> Clearing request in cache error: %v", requestID, err)
	}

	return err
}
