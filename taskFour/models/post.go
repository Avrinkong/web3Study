package models

import (
	"time"
)

type Post struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Title     string    `gorm:"not null;size:200" json:"title"`
	Content   string    `gorm:"type:text;not null" json:"content"`
	UserID    uint      `gorm:"not null" json:"user_id"`
	User      User      `gorm:"foreignKey:UserID" json:"user"`
	Comments  []Comment `gorm:"foreignKey:PostID;constraint:OnDelete:CASCADE" json:"comments,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
