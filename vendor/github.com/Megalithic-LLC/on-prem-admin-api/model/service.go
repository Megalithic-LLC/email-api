package model

import (
	"time"
)

type Service struct {
	ID              string     `json:"id" gorm:"primary_key;type:char(20)"`
	Name            string     `json:"name" gorm:"size:50;not null;unique_index"`
	DisplayName     string     `json:"displayName" gorm:"size:50"`
	Description     string     `json:"description"`
	LongDescription string     `json:"longDescription" gorm:"type:text"`
	ImageUrl        string     `json:"imageUrl"`
	Visible         bool       `json:"visible"`
	CreatedAt       time.Time  `json:"createdAt"`
	UpdatedAt       time.Time  `json:"updatedAt"`
	DeletedAt       *time.Time `json:"deletedAt" gorm:"index"`
}
