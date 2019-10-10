package model

import (
	"time"
)

type Plan struct {
	ID          string     `json:"id" gorm:"primary_key;type:char(20)"`
	ServiceID   string     `json:"service" gorm:"type:char(20);index"`
	Name        string     `json:"name" gorm:"size:50;not null"`
	DisplayName string     `json:"displayName" gorm:"size:50"`
	Description string     `json:"description"`
	Free        bool       `json:"free"`
	Visible     bool       `json:"visible"`
	CreatedAt   time.Time  `json:"createdAt"`
	UpdatedAt   time.Time  `json:"updatedAt"`
	DeletedAt   *time.Time `json:"deletedAt" gorm:"index"`
}
