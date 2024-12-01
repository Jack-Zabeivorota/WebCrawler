package database

import (
	dbm "main/database/models"
	"main/models"
)

type MockDataBase struct {
	count int64
}

func (m *MockDataBase) AddRequest(data *models.RequestData, startUrl string) (int64, error) {
	m.count++
	println("Added request ", m.count, " for URL ", startUrl)
	return m.count, nil
}

func (MockDataBase) GetRequest(ID int64) (*dbm.Request, error) {
	println("Geted request ", ID)
	return &dbm.Request{}, nil
}

func (MockDataBase) DeleteRequest(ID int64) (bool, error) {
	println("Deleted request ", ID)
	return true, nil
}

func (MockDataBase) GetURLsFromRequest(requestID int64) ([]dbm.URL, error) {
	println("Geted URLs request ", requestID)
	return []dbm.URL{{}, {}}, nil
}

func (MockDataBase) DeleteRequestAndURLs(requestID int64) (bool, error) {
	println("Deleted request ", requestID, " and URLs")
	return true, nil
}
