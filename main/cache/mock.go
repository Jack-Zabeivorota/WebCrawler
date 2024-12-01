package cache

import "main/models"

type MockCache struct{}

func (MockCache) SetRequestData(requestID int64, data *models.RequestData) error {
	print("Set request: ", requestID, " SameDomain: ", data.SameDomainOnly, " words: ", data.Words, "\n")
	return nil
}

func (MockCache) SetURLToAllURLs(requestID int64, url string) error {
	print("Set ", url, " from ", requestID, " to InWork\n")
	return nil
}

func (MockCache) ClearRequest(requestID int64) error {
	print("Clear all data for request ", requestID, "\n")
	return nil
}
