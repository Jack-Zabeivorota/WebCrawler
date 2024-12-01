package cache

import "main/models"

type ICache interface {
	SetRequestData(requestID int64, data *models.RequestData) error
	SetURLToAllURLs(requestID int64, url string) error
	ClearRequest(requestID int64) error
}
