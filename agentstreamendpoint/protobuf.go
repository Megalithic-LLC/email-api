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
