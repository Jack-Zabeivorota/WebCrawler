package models

type URL struct {
	ID          int64  `gorm:"primaryKey"`
	URL         string `gorm:"not null"`
	RequestID   int64  `gorm:"foreignKey;not null"`
	Status      int    `gorm:"not null;type:smallint"`
	FindedWords string `gorm:"not null"`
}

var URLStatus = struct {
	Success, Unreaded, Fail int
}{
	Success:  0,
	Fail:     1,
	Unreaded: 2,
}
