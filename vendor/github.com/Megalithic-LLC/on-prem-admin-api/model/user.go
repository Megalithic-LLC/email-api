package model

import (
	"time"
)

type User struct {
	ID        string `gorm:"primary_key;type:char(20)"`
	Username  string `gorm:"size:30;unique_index;not null"`
	First     string `gorm:"size:50"`
	Last      string `gorm:"size:50"`
	Password  []byte `gorm:"size:60"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time `sql:"index"`
}
