package model

import (
	"crypto/md5"
	"fmt"
	"time"
)

type Snapshot struct {
	ID              string     `json:"id" gorm:"primary_key;type:char(20)"`
	AgentID         string     `json:"agent" gorm:"type:char(20);index"`
	Name            string     `json:"name" gorm:"type:varchar(100);index"`
	Engine          string     `json:"engine" gorm:"size:25"`
	Progress        float32    `json:"progress"`
	Size            uint64     `json:"size"`
	CreatedByUserID string     `json:"createdBy" gorm:"type:char(20)"`
	CreatedAt       time.Time  `json:"createdAt"`
	UpdatedAt       time.Time  `json:"updatedAt"`
	DeletedAt       *time.Time `json:"deletedAt" gorm:"index"`
}

func (self Snapshot) Hash() []byte {
	hasher := md5.New()
	hasher.Write([]byte(self.ID))
	hasher.Write([]byte(self.AgentID))
	hasher.Write([]byte(self.Name))
	hasher.Write([]byte(self.Engine))
	hasher.Write([]byte(fmt.Sprintf("%v", self.Size)))
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
