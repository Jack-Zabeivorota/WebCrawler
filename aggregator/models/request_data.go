package models

type RequestData struct {
	SameDomainOnly bool     `json:"domain_only"`
	Words          []string `json:"words"`
}
