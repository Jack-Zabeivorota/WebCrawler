package models

type SignData struct {
	Password string         `json:"password"`
	Sign     string         `json:"sign"`
	Services map[string]int `json:"services"`
}
