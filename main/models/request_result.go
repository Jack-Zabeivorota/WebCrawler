package models

type RequestResult struct {
	ID             int64       `json:"id"`
	StartURL       string      `json:"start_url"`
	Words          []string    `json:"words"`
	SameDomainOnly bool        `json:"same_domain_only"`
	URLs           []URLResult `json:"urls"`
}
