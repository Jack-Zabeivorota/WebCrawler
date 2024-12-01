package models

type URLData struct {
	URL            string   `json:"url"`
	SameDomainOnly bool     `json:"same_domain_only"`
	Words          []string `json:"words"`
}
