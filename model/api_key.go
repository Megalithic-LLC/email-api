package model

import (
	"crypto/md5"
	"time"
)

type ApiKey struct {
	ID              string     `json:"id" gorm:"primary_key;type:char(20)"`
	Key             string     `json:"key" gorm:"size:20;index"`
	Description     string     `json:"description" gorm:"size:255"`
	CreatedByUserID string     `json:"createdBy" gorm:"type:char(20)"`
	CreatedAt       time.Time  `json:"createdAt"`
	UpdatedAt       time.Time  `json:"updatedAt"`
	DeletedAt       *time.Time `json:"deletedAt" gorm:"index"`
}

func (self ApiKey) Hash() []byte {
	hasher := md5.New()
	hasher.Write([]byte(self.ID))
	hasher.Write([]byte(self.Key))
	hasher.Write([]byte(self.Description))
	hasher.Write([]byte(self.CreatedByUserID))

	createdAtAsBinary, _ := self.CreatedAt.MarshalBinary()
	hasher.Write(createdAtAsBinary)

	updatedAtAsBinary, _ := self.UpdatedAt.MarshalBinary()
	hasher.Write(updatedAtAsBinary)

	if self.DeletedAt != nil {
		deletedAtAsBinary, _ := self.DeletedAt.MarshalBinary()
		hasher.Write(deletedAtAsBinary)
	}

	return hasher.Sum(nil)
}
