package agentstreamendpoint

import (
	"github.com/on-prem-net/email-api/agentstreamendpoint/emailproto"
	"github.com/on-prem-net/email-api/model"
)

func (self *AgentStream) SendGetAccountsResponse(requestId uint64, accounts []model.Account) error {
	res := emailproto.ServerMessage{
		Id: requestId,
		MessageType: &emailproto.ServerMessage_GetAccountsResponse{
			GetAccountsResponse: &emailproto.GetAccountsResponse{
				Accounts: AccountsAsProtobuf(accounts),
			},
		},
	}
	return self.SendResponse(res)
}
