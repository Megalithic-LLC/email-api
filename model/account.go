package model

import (
	"crypto/md5"
	"time"
)

type Account struct {
	ID                string     `json:"id" gorm:"primary_key;type:char(20)"`
	AgentID           string     `json:"agent" gorm:"type:char(20);index"`
	ServiceInstanceID string     `json:"serviceInstance" gorm:"type:char(20);index"`
	Name              string     `json:"name" gorm:"size:100;index"`
	DomainID          string     `json:"domain" gorm:"type:char(20);index"`
	Email             string     `json:"email" gorm:"size:255;unique_index"`
	First             string     `json:"first" gorm:"size:50"`
	Last              string     `json:"last" gorm:"size:50"`
	DisplayName       string     `json:"displayName" gorm:"size:100"`
	Password          []byte     `json:"password" gorm:"size:50"`
	CreatedByUserID   string     `json:"createdBy" gorm:"type:char(20)"`
	CreatedAt         time.Time  `json:"createdAt"`
	UpdatedAt         time.Time  `json:"updatedAt"`
	DeletedAt         *time.Time `json:"deletedAt" gorm:"index"`
}

func (self Account) Hash() []byte {
	hasher := md5.New()
	hasher.Write([]byte(self.ID))
	hasher.Write([]byte(self.AgentID))
	hasher.Write([]byte(self.ServiceInstanceID))
	hasher.Write([]byte(self.Name))
	hasher.Write([]byte(self.DomainID))
	hasher.Write([]byte(self.Email))
	hasher.Write([]byte(self.First))
	hasher.Write([]byte(self.Last))
	hasher.Write([]byte(self.DisplayName))
	hasher.Write(self.Password)
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
