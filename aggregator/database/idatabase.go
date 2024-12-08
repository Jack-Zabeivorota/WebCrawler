package database

type IDataBase interface {
	AddRequestResults(requestID int64, urlResults map[string]string) error
}
