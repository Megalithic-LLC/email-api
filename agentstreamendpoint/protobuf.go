package agentstreamendpoint

import (
	"github.com/on-prem-net/email-api/agentstreamendpoint/emailproto"
	"github.com/on-prem-net/email-api/model"
)

func AccountAsProtobuf(account model.Account) emailproto.Account {
	return emailproto.Account{
		Id:          account.ID,
		Name:        account.Name,
		DomainId:    account.DomainID,
		Email:       account.Email,
		First:       account.First,
		Last:        account.Last,
		DisplayName: account.DisplayName,
		Password:    account.Password,
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

func DomainAsProtobuf(domain model.Domain) emailproto.Domain {
	return emailproto.Domain{
		Id:   domain.ID,
		Name: domain.Name,
	}
}

func DomainsAsProtobuf(domains []model.Domain) []*emailproto.Domain {
	pbDomains := []*emailproto.Domain{}
	for _, domain := range domains {
		pbDomain := DomainAsProtobuf(domain)
		pbDomains = append(pbDomains, &pbDomain)
	}
	return pbDomains
}

func EndpointAsProtobuf(endpoint model.Endpoint) emailproto.Endpoint {
	return emailproto.Endpoint{
		Id:       endpoint.ID,
		Protocol: endpoint.Protocol,
		Type:     endpoint.Type,
		Port:     uint32(endpoint.Port),
		Path:     endpoint.Path,
		Enabled:  endpoint.Enabled,
	}
}

func EndpointsAsProtobuf(endpoints []model.Endpoint) []*emailproto.Endpoint {
	pbEndpoints := []*emailproto.Endpoint{}
	for _, endpoint := range endpoints {
		pbEndpoint := EndpointAsProtobuf(endpoint)
		pbEndpoints = append(pbEndpoints, &pbEndpoint)
	}
	return pbEndpoints
}

func SnapshotAsProtobuf(snapshot model.Snapshot) emailproto.Snapshot {
	return emailproto.Snapshot{
		Id:        snapshot.ID,
		ServiceId: snapshot.ServiceID,
		Name:      snapshot.Name,
		Engine:    snapshot.Engine,
		Progress:  snapshot.Progress,
		Size:      snapshot.Size,
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

func SnapshotFromProtobuf(pbSnapshot *emailproto.Snapshot) *model.Snapshot {
	if pbSnapshot == nil {
		return nil
	}
	return &model.Snapshot{
		ID:        pbSnapshot.Id,
		ServiceID: pbSnapshot.ServiceId,
		Name:      pbSnapshot.Name,
		Engine:    pbSnapshot.Engine,
		Progress:  pbSnapshot.Progress,
		Size:      pbSnapshot.Size,
	}
}
