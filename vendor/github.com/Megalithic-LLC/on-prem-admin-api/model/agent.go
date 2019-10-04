package model

import (
	"time"
)

type Agent struct {
	ID                 string     `json:"id" gorm:"primary_key;type:char(20)"`
	OwnerUserID        string     `json:"owner" gorm:"type:char(20);index"`
	ServiceInstanceIDs []string   `json:"serviceInstances" gorm:"-"`
	SnapshotIDs        []string   `json:"snapshots" gorm:"-"`
	CreatedAt          time.Time  `json:"createdAt"`
	UpdatedAt          time.Time  `json:"updatedAt"`
	DeletedAt          *time.Time `json:"deletedAt" gorm:"index"`
}
