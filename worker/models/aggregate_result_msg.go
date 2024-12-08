package models

const AggregateResultTopic = "AggregateResult"

type AggregateResultMsg struct {
	RequestID int64 `json:"request_id"`
}
