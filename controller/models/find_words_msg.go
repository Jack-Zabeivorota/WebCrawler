package models

const FindWordsTopic = "FindWords"

type FindWordsMsg struct {
	RequestID int64  `json:"request_id"`
	URL       string `json:"url"`
	Attempts  int    `json:"attempts"`
}
