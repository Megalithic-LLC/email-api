package model

import (
	"time"
)

type Agent struct {
	ID                 string     `json:"id" gorm:"primary_key;type:char(20)"`
	ServiceInstanceIDs []string   `json:"serviceInstances" gorm:"-"`
	SnapshotIDs        []string   `json:"snapshots" gorm:"-"`
	CreatedByUserID    string     `json:"createdBy" gorm:"type:char(20)"`
	CreatedAt          time.Time  `json:"createdAt"`
	UpdatedAt          time.Time  `json:"updatedAt"`
	DeletedAt          *time.Time `json:"deletedAt" gorm:"index"`
}
