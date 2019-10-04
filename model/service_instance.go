package model

import (
	"time"
)

type ServiceInstance struct {
	ID        string     `json:"id" gorm:"primary_key;type:char(20)"`
	AgentID   string     `json:"agent" gorm:"type:char(20);index"`
	ServiceID string     `json:"service" gorm:"type:char(20);index"`
	PlanID    string     `json:"plan" gorm:"type:char(20);index"`
	CreatedAt time.Time  `json:"createdAt"`
	UpdatedAt time.Time  `json:"updatedAt"`
	DeletedAt *time.Time `json:"deletedAt" gorm:"index"`
}
