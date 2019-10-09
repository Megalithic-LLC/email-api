package model

import (
	"time"
)

type SnapshotChunk struct {
	ID         uint64     `json:"id" gorm:"primary_key;auto_increment"`
	SnapshotID string     `json:"snapshotId" gorm:"type:char(20);index"`
	Number     uint32     `json:"number"`
	Data       []byte     `json:"data" gorm:"size:1000000"`
	CreatedAt  time.Time  `json:"createdAt"`
	DeletedAt  *time.Time `json:"deletedAt" gorm:"index"`
}
