package models

import (
	"time"

	"gorm.io/gorm"
)

// Post represents a single anonymous confession.
type Post struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	Content   string         `gorm:"not null" json:"content"`
	Score     int            `gorm:"not null;default:0" json:"score"`
	Hidden    bool           `gorm:"not null;default:false" json:"-"` // Hidden from API responses
	CreatedAt time.Time      `json:"createdAt"`
	UpdatedAt time.Time      `json:"updatedAt"`
	Votes     []Vote         `gorm:"foreignKey:PostID" json:"-"` // Has-many relationship
}

// Vote represents a +1 or -1 vote on a Post.
type Vote struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	PostID    uint           `gorm:"not null;index" json:"postId"`
	Value     int            `gorm:"not null" json:"value"` // Should be +1 or -1
	CreatedAt time.Time      `json:"createdAt"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}