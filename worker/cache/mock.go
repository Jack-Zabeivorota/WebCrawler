package cache

import "main/models"

type MockCache struct{}

func (MockCache) SetURLsToAllURLs(requestID int64, urls []string) error {
	print("SET ID: ", requestID, " URLs: ", urls, "\n")
	return nil
}

func (MockCache) SetURLToCompleteds(requestID int64, url, status string, words []string) error {
	print("SET ID: ", requestID, " URLs: ", url, " Status: ", status, " Words: ", words, "\n")
	return nil
}

func (MockCache) URLIsCompleted(requestID int64, url string) (bool, error) {
	print("Exists ", url, " in ", requestID, "\n")
	return true, nil
}

func (MockCache) GetNotCompletedURLs(requestID int64, urls []string) ([]string, error) {
	print("ID: ", requestID, " Urls: ", urls, "\n")
	return []string{}, nil
}

func (MockCache) GetRequestData(requestID int64) (*models.RequestData, error) {
	print("Get data from ", requestID, "\n")
	return &models.RequestData{}, nil
}
