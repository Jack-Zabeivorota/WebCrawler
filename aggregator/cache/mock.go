package cache

type MockCache struct{}

func (MockCache) AllURLsIsCompleted(requestID int64) (bool, error) {
	print("SET ID: ", requestID, "\n")
	return false, nil
}

func (MockCache) GetURLsResult(requestID int64) (map[string]string, error) {
	print("Get data from ", requestID, "\n")
	return map[string]string{}, nil
}

func (MockCache) ClearRequest(requestID int64) error {
	print("Clear ", requestID, "\n")
	return nil
}
