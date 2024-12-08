package database

import (
	dbm "main/database/models"
	"main/models"
)

type IDataBase interface {
	AddRequest(data *models.RequestData, startUrl string) (int64, error)
	GetRequest(ID int64) (*dbm.Request, error)
	DeleteRequest(ID int64) (bool, error)
	GetURLsFromRequest(requestID int64) ([]dbm.URL, error)
	DeleteRequestAndURLs(requestID int64) (bool, error)
}
