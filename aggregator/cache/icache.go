package cache

type ICache interface {
	AllURLsIsCompleted(requestID int64) (bool, error)
	GetURLsResult(requestID int64) (map[string]string, error)
	ClearRequest(requestID int64) error
}
