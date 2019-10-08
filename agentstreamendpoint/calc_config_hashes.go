package agentstreamendpoint

import (
	"crypto/md5"

	"github.com/on-prem-net/email-api/model"
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

	// serviceInstances
	{
		serviceInstances := []model.ServiceInstance{}
		searchFor := &model.ServiceInstance{AgentID: self.agentID}
		if err := self.endpoint.db.Where(searchFor).Find(&serviceInstances).Error; err != nil {
			return nil, err
		}
		hasher := md5.New()
		for _, serviceInstance := range serviceInstances {
			hasher.Write(serviceInstance.Hash())
		}
		hashesByTable["serviceInstances"] = hasher.Sum(nil)
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
