package cache

import "main/models"

type ICache interface {
	SetURLsToAllURLs(requestID int64, urls []string) error
	SetURLToCompleteds(requestID int64, url, status string, words []string) error
	URLIsCompleted(requestID int64, url string) (bool, error)
	GetNotProcessedURLs(requestID int64, urls []string) ([]string, error)
	GetRequestData(requestID int64) (*models.RequestData, error)
}
