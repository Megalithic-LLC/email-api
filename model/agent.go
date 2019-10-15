package model

import (
	"time"
)

type Agent struct {
	ID              string     `json:"id" gorm:"primary_key;type:char(20)"`
	PlanID          string     `json:"plan" gorm:"type:char(20)"`
	AccountIDs      []string   `json:"accounts" gorm:"-"`
	DomainIDs       []string   `json:"domains" gorm:"-"`
	EndpointIDs     []string   `json:"endpoints" gorm:"-"`
	SnapshotIDs     []string   `json:"snapshots" gorm:"-"`
	CreatedByUserID string     `json:"createdBy" gorm:"type:char(20)"`
	CreatedAt       time.Time  `json:"createdAt"`
	UpdatedAt       time.Time  `json:"updatedAt"`
	DeletedAt       *time.Time `json:"deletedAt" gorm:"index"`
}
