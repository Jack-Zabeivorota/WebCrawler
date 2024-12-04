package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/redis/go-redis/v9"

	"main/logger"
	"main/models"
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

func (rc *RedisCache) SetURLsToAllURLs(requestID int64, urls []string) error {
	if len(urls) == 0 {
		return nil
	}

	ctx := context.Background()
	id := fmt.Sprintf("all_urls:%d", requestID)

	anyUrls := tools.Select(urls, func(url string) any {
		return url
	})

	err := tools.RetryCycle(func() error {
		return rc.client.SAdd(ctx, id, anyUrls...).Err()
	}, "Redis fail try", false)

	if err != nil {
		rc.log.Error("Setting URL to all URLs error: %v", err)
	}

	return err
}

func (rc *RedisCache) SetURLToCompleteds(requestID int64, url, status string, words []string) error {
	ctx := context.Background()
	id := fmt.Sprintf("completed_urls:%d", requestID)

	words = append(words, status)
	value := strings.Join(words, ",")

	err := tools.RetryCycle(func() error {
		return rc.client.HSet(ctx, id, url, value).Err()
	}, "Redis fail try", false)

	if err != nil {
		rc.log.Error("Setting URL to completeds error: %v", err)
	}

	return err
}

func (rc *RedisCache) URLIsCompleted(requestID int64, url string) (bool, error) {
	ctx := context.Background()
	id := fmt.Sprintf("completed_urls:%d", requestID)

	var exists bool
	var err error

	tools.RetryCycle(func() error {
		exists, err = rc.client.HExists(ctx, id, url).Result()
		return err
	}, "Redis fail try", false)

	if err != nil {
		rc.log.Error("URL is completed checking error: %v", err)
	}

	return exists, err
}

func (rc *RedisCache) GetNotProcessedURLs(requestID int64, urls []string) ([]string, error) {
	ctx := context.Background()
	id := fmt.Sprintf("all_urls:%d", requestID)
	notExistsUrls := []string{}

	var exists bool
	var err error

	for _, url := range urls {
		tools.RetryCycle(func() error {
			exists, err = rc.client.SIsMember(ctx, id, url).Result()
			return err
		}, "Redis fail try", false)

		if err != nil {
			rc.log.Error("Getting not found in all_urls error: %v", err)
			return []string{}, err
		}

		if !exists {
			notExistsUrls = append(notExistsUrls, url)
		}
	}

	return notExistsUrls, nil
}

func (rc *RedisCache) GetRequestData(requestID int64) (*models.RequestData, error) {
	ctx := context.Background()
	id := fmt.Sprintf("%d", requestID)

	var data []byte
	var err error

	tools.RetryCycle(func() error {
		data, err = rc.client.HGet(ctx, "requests_data", id).Bytes()
		return err
	}, "Redis fail try", false)

	if err != nil {
		rc.log.Error("Getting request data error: %v", err)
		return nil, err
	}

	requestData := &models.RequestData{}
	err = json.Unmarshal(data, requestData)

	if err != nil {
		rc.log.Error("Getting request data error: %v", err)
	}

	return requestData, err
}
