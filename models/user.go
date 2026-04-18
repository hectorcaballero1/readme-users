package models

import "time"

type User struct {
	ID           uint      `json:"id"         gorm:"primaryKey"`
	Name         string    `json:"name"       gorm:"size:100;not null"`
	Email        string    `json:"email"      gorm:"size:150;uniqueIndex;not null"`
	PasswordHash string    `json:"-"          gorm:"column:password_hash;size:255;not null"`
	ZoneID       *uint     `json:"zone_id"`
	Zone         *Zone     `json:"zone,omitempty" gorm:"foreignKey:ZoneID"`
	PhotoURL     *string   `json:"photo_url"`
	CreatedAt    time.Time `json:"created_at"`
}

// UserResponse is returned in all API responses — never includes PasswordHash.
type UserResponse struct {
	ID        uint      `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	ZoneID    *uint     `json:"zone_id"`
	Zone      *Zone     `json:"zone,omitempty"`
	PhotoURL  *string   `json:"photo_url"`
	CreatedAt time.Time `json:"created_at"`
}
