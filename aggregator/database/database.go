package database

import (
	"errors"
	"strings"

	dbm "main/database/models"
	"main/logger"
	"main/tools"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type DataBase struct {
	client *gorm.DB
	log    *logger.Logger
}

func NewPostgreSQL(dsn string) *DataBase {
	log := logger.Instance()

	if dsn == "" {
		dsn = "host=localhost port=5432 user=postgres password=admin dbname=WebCrawler sslmode=disable"
	}

	var client *gorm.DB
	var err error

	tools.RetryCycle(func() error {
		client, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
		return err
	}, "PostgreSQL fail connecting try", true)

	return &DataBase{
		client: client,
		log:    log,
	}
}

func (db *DataBase) Init() error {
	err := tools.RetryCycle(func() error {
		return db.client.AutoMigrate(
			&dbm.Request{},
			&dbm.URL{},
		)
	}, "PostgreSQL fail init try", false)

	if err != nil {
		db.log.Error("Auto migrate error: %v", err)
	}

	return err
}

func (db *DataBase) getStatusAndWords(source string) (int, string, error) {
	index := strings.LastIndex(source, ",")

	var status, words string

	if index == -1 {
		status = source
	} else {
		words = source[:index]
		status = source[index+1:]
	}

	statusDict := map[string]int{
		"success":  dbm.URLStatus.Success,
		"fail":     dbm.URLStatus.Fail,
		"unreaded": dbm.URLStatus.Unreaded,
	}

	intStatus, ok := statusDict[status]

	if !ok {
		db.log.Error("Incorrect URL status: %s", status)
		return 0, "", errors.New("incorrect URL status")
	}

	return intStatus, words, nil
}

func (db *DataBase) AddRequestResults(requestID int64, urlResults map[string]string) error {
	urls := make([]dbm.URL, 0, len(urlResults))

	for url, value := range urlResults {
		status, words, err := db.getStatusAndWords(value)

		if err != nil {
			return err
		}

		urls = append(urls, dbm.URL{
			URL:         url,
			RequestID:   requestID,
			Status:      status,
			FindedWords: words,
		})
	}

	err := tools.RetryCycle(func() error {
		return db.client.Transaction(func(tx *gorm.DB) error {

			err := tx.Model(&dbm.Request{}).
				Where("id = ?", requestID).
				Update("is_done", true).Error

			if err != nil {
				return err
			}

			if len(urls) == 0 {
				return nil
			}

			return tx.Create(urls).Error

		})
	}, "PostgreSQL fail try", false)

	if err != nil {
		db.log.Error("Request %d -> Additing request results error: %v", requestID, err)
	}

	return err
}
