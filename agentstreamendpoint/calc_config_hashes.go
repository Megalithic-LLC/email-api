package agentstreamendpoint

import (
	"crypto/md5"

	"github.com/Megalithic-LLC/on-prem-email-api/model"
)

func (self *AgentStream) calcConfigHashes() (map[string][]byte, error) {

	hashesByTable := map[string][]byte{}

	// accounts
	{
		accounts := []model.Account{}
		searchFor := &model.Account{AgentID: self.agentID}
		if err := self.endpoint.db.Where(searchFor).Find(&accounts).Error; err != nil {
			return nil, err
		}
		hasher := md5.New()
		for _, account := range accounts {
			hasher.Write(account.Hash())
		}
		hashesByTable["accounts"] = hasher.Sum(nil)
	}

	// domains
	{
		domains := []model.Domain{}
		searchFor := &model.Domain{AgentID: self.agentID}
		if err := self.endpoint.db.Where(searchFor).Find(&domains).Error; err != nil {
			return nil, err
		}
		hasher := md5.New()
		for _, domain := range domains {
			hasher.Write(domain.Hash())
		}
		hashesByTable["domains"] = hasher.Sum(nil)
	}

	// endpoints
	{
		endpoints := []model.Endpoint{}
		searchFor := &model.Endpoint{AgentID: self.agentID}
		if err := self.endpoint.db.Where(searchFor).Find(&endpoints).Error; err != nil {
			return nil, err
		}
		hasher := md5.New()
		for _, endpoint := range endpoints {
			hasher.Write(endpoint.Hash())
		}
		hashesByTable["endpoints"] = hasher.Sum(nil)
	}

	// snapshots
	{
		snapshots := []model.Snapshot{}
		searchFor := &model.Snapshot{AgentID: self.agentID}
		if err := self.endpoint.db.Where(searchFor).Find(&snapshots).Error; err != nil {
			return nil, err
		}
		hasher := md5.New()
		for _, snapshot := range snapshots {
			hasher.Write(snapshot.Hash())
		}
		hashesByTable["snapshots"] = hasher.Sum(nil)
	}

	return hashesByTable, nil
}
