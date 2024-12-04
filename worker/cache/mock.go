package cache

import "main/models"

type MockCache struct{}

func (MockCache) SetURLsToAllURLs(requestID int64, urls []string) error {
	println("SET URLs: ", urls, " in all_urls:", requestID)
	return nil
}

func (MockCache) SetURLToCompleteds(requestID int64, url, status string, words []string) error {
	println("SET URL: ", url, " in completed_urls:", requestID, " with status: ", status, " and words: ", words)
	return nil
}

func (MockCache) URLIsCompleted(requestID int64, url string) (bool, error) {
	println("Exists ", url, " in completed_urls:", requestID)
	return true, nil
}

func (MockCache) GetNotProcessedURLs(requestID int64, urls []string) ([]string, error) {
	println("GET not found URLs in all_urls:", requestID)
	return []string{}, nil
}

func (MockCache) GetRequestData(requestID int64) (*models.RequestData, error) {
	println("Get data from requests_data ", requestID)
	return &models.RequestData{}, nil
}
