package agentstreamendpoint

import (
	"github.com/on-prem-net/email-api/agentstreamendpoint/emailproto"
	"github.com/on-prem-net/email-api/model"
)

func AccountAsProtobuf(account model.Account) emailproto.Account {
	return emailproto.Account{
		Id:                account.ID,
		ServiceInstanceId: account.ServiceInstanceID,
		Name:              account.Name,
		Email:             account.Email,
		First:             account.First,
		Last:              account.Last,
		DisplayName:       account.DisplayName,
		Password:          account.Password,
	}
}

func AccountsAsProtobuf(accounts []model.Account) []*emailproto.Account {
	pbAccounts := []*emailproto.Account{}
	for _, account := range accounts {
		pbAccount := AccountAsProtobuf(account)
		pbAccounts = append(pbAccounts, &pbAccount)
	}
	return pbAccounts
}

func ServiceInstanceAsProtobuf(serviceInstance model.ServiceInstance) emailproto.ServiceInstance {
	return emailproto.ServiceInstance{
		Id:        serviceInstance.ID,
		ServiceId: serviceInstance.ServiceID,
		PlanId:    serviceInstance.PlanID,
	}
}

func ServiceInstancesAsProtobuf(serviceInstances []model.ServiceInstance) []*emailproto.ServiceInstance {
	pbServiceInstances := []*emailproto.ServiceInstance{}
	for _, serviceInstance := range serviceInstances {
		pbServiceInstance := ServiceInstanceAsProtobuf(serviceInstance)
		pbServiceInstances = append(pbServiceInstances, &pbServiceInstance)
	}
	return pbServiceInstances
}

func SnapshotAsProtobuf(snapshot model.Snapshot) emailproto.Snapshot {
	return emailproto.Snapshot{
		Id:   snapshot.ID,
		Name: snapshot.Name,
	}
}

func SnapshotsAsProtobuf(snapshots []model.Snapshot) []*emailproto.Snapshot {
	pbSnapshots := []*emailproto.Snapshot{}
	for _, snapshot := range snapshots {
		pbSnapshot := SnapshotAsProtobuf(snapshot)
		pbSnapshots = append(pbSnapshots, &pbSnapshot)
	}
	return pbSnapshots
}
