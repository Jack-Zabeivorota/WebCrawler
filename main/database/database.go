package database

import (
	"strings"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	dbm "main/database/models"
	"main/logger"
	"main/models"
	"main/tools"
)

type DataBase struct {
	client *gorm.DB
	log    *logger.Logger
}

func NewPostgreSQL(dsn string) *DataBase {
	if dsn == "" {
		dsn = "host=localhost port=5432 user=postgres password=admin dbname=WebCrawler sslmode=disable"
	}

	var client *gorm.DB
	var err error

	tools.RetryCycle(func() error {
		client, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
		return err
	}, "PostgreSQL fail connecting try", true)

	return &DataBase{client: client}
}

func (db *DataBase) Init() error {
	err := tools.RetryCycle(func() error {
		return db.client.AutoMigrate(
			&dbm.Request{},
			&dbm.URL{},
		)
	}, "PostgreSQL fail init try", false)

	if err != nil {
		db.log.Error("DB initiating error: %s", err.Error())
	}

	return err
}

func (db *DataBase) AddRequest(data *models.RequestData, startUrl string) (int64, error) {
	request := &dbm.Request{
		StartURL:       startUrl,
		Words:          strings.Join(data.Words, ","),
		SameDomainOnly: data.SameDomainOnly,
	}

	err := db.client.Create(request).Error

	if err != nil {
		db.log.Error("Additing request to DB error: %s", err.Error())
	}

	return request.ID, err
}

func (db *DataBase) GetRequest(ID int64) (*dbm.Request, error) {
	request := &dbm.Request{}
	err := db.client.First(request, "id = ?", ID).Error

	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}

	if err != nil {
		db.log.Error("Getting request from DB error: %v", err)
	}

	return request, err
}

func (db *DataBase) DeleteRequest(ID int64) (bool, error) {
	result := db.client.Delete(&dbm.Request{}, ID)

	if result.Error != nil {
		db.log.Error("Deleting request error: %v", result.Error)
	}

	return result.RowsAffected > 0, result.Error
}

func (db *DataBase) GetURLsFromRequest(requestID int64) ([]dbm.URL, error) {
	urls := []dbm.URL{}
	err := db.client.Find(&urls, "request_id = ?", requestID).Error

	if err != nil {
		db.log.Error("Getting URLs from request error: %v", err)
	}

	return urls, err
}

func (db *DataBase) DeleteRequestAndURLs(requestID int64) (bool, error) {
	isDeleted := false

	err := db.client.Transaction(func(tx *gorm.DB) error {

		result := tx.Delete(&dbm.Request{}, requestID)

		if result.Error != nil {
			return result.Error
		}
		isDeleted = result.RowsAffected > 0

		result = tx.Delete(&dbm.URL{}, "request_id = ?", requestID)

		if result.Error != nil {
			return result.Error
		}
		isDeleted = isDeleted || result.RowsAffected > 0

		return nil

	})

	if err != nil {
		db.log.Error("Deleting request and URLs error: %v", err)
	}

	return isDeleted, err
}
