package models

type Request struct {
	ID             int64  `gorm:"primaryKey"`
	StartURL       string `gorm:"not null"`
	Words          string `gorm:"not null"`
	SameDomainOnly bool   `gorm:"not null"`
	IsDone         bool   `gorm:"not null"`
}
