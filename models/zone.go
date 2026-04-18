package models

type Zone struct {
	ID   uint   `json:"id"   gorm:"primaryKey"`
	Name string `json:"name" gorm:"size:100;not null"`
}
