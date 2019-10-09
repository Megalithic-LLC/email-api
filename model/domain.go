package model

import (
	"crypto/md5"
	"time"
)

type Domain struct {
	ID                string     `json:"id" gorm:"primary_key;type:char(20)"`
	AgentID           string     `json:"agent" gorm:"type:char(20);index"`
	ServiceInstanceID string     `json:"serviceInstance" gorm:"type:char(20);index"`
	Name              string     `json:"name" gorm:"size:253;index"`
	CreatedByUserID   string     `json:"createdBy" gorm:"type:char(20)"`
	CreatedAt         time.Time  `json:"createdAt"`
	UpdatedAt         time.Time  `json:"updatedAt"`
	DeletedAt         *time.Time `json:"deletedAt" gorm:"index"`
}

func (self Domain) Hash() []byte {
	hasher := md5.New()
	hasher.Write([]byte(self.ID))
	hasher.Write([]byte(self.AgentID))
	hasher.Write([]byte(self.ServiceInstanceID))
	hasher.Write([]byte(self.Name))
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
