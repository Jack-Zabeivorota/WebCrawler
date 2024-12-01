package database

type MockDataBase struct{}

func (MockDataBase) AddRequestResults(requestID int64, urlResults map[string]string) error {
	println("Sent results for request ", requestID, ": ", urlResults)
	return nil
}
