package model

import (
	"crypto/md5"
	"time"
)

type ServiceInstance struct {
	ID         string     `json:"id" gorm:"primary_key;type:char(20)"`
	AgentID    string     `json:"agent" gorm:"type:char(20);index"`
	ServiceID  string     `json:"service" gorm:"type:char(20);index"`
	AccountIDs []string   `json:"accounts" gorm:"-"`
	PlanID     string     `json:"plan" gorm:"type:char(20);index"`
	CreatedAt  time.Time  `json:"createdAt"`
	UpdatedAt  time.Time  `json:"updatedAt"`
	DeletedAt  *time.Time `json:"deletedAt" gorm:"index"`
}

func (self ServiceInstance) Hash() []byte {
	hasher := md5.New()
	hasher.Write([]byte(self.ID))
	hasher.Write([]byte(self.AgentID))
	hasher.Write([]byte(self.ServiceID))
	for _, accountID := range self.AccountIDs {
		hasher.Write([]byte(accountID))
	}
	hasher.Write([]byte(self.PlanID))

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
