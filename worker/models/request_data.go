package models

type RequestData struct {
	Words          []string `json:"words"`
	SameDomainOnly bool     `json:"same_domain_only"`
}
