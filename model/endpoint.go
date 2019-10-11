package model

import (
	"crypto/md5"
	"fmt"
	"time"
)

type Endpoint struct {
	ID              string     `json:"id" gorm:"primary_key;type:char(20)"`
	AgentID         string     `json:"agent" gorm:"type:char(20);index"`
	Protocol        string     `json:"protocol" gorm:"size:25"`
	Type            string     `json:"type" gorm:"size:25"`
	Port            uint16     `json:"port"`
	Path            string     `json:"path" gorm:"size:255"`
	Enabled         bool       `json:"enabled"`
	CreatedByUserID string     `json:"createdBy" gorm:"type:char(20)"`
	CreatedAt       time.Time  `json:"createdAt"`
	UpdatedAt       time.Time  `json:"updatedAt"`
	DeletedAt       *time.Time `json:"deletedAt" gorm:"index"`
}

func (self Endpoint) Hash() []byte {
	hasher := md5.New()
	hasher.Write([]byte(self.ID))
	hasher.Write([]byte(self.AgentID))
	hasher.Write([]byte(self.Protocol))
	hasher.Write([]byte(self.Type))
	hasher.Write([]byte(fmt.Sprintf("%v", self.Port)))
	hasher.Write([]byte(self.Path))
	hasher.Write([]byte(fmt.Sprintf("%v", self.Enabled)))
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
