package models

type URLResult struct {
	URL         string   `json:"url"`
	Status      string   `json:"status"`
	FindedWords []string `json:"finded_words"`
}
