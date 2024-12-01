package models

type RequestData struct {
	Words          []string `json:"words"`
	SameDomainOnly bool     `json:"domain_only"`
}
