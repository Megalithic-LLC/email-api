package agentstreamendpoint

import (
	"github.com/on-prem-net/email-api/agentstreamendpoint/emailproto"
	"github.com/on-prem-net/email-api/model"
)

func (self *AgentStream) SendGetDomainsResponse(requestId uint64, domains []model.Domain) error {
	getDomainsRes := emailproto.ServerMessage{
		Id: requestId,
		MessageType: &emailproto.ServerMessage_GetDomainsResponse{
			GetDomainsResponse: &emailproto.GetDomainsResponse{
				Domains: DomainsAsProtobuf(domains),
			},
		},
	}
	return self.SendResponse(getDomainsRes)
}
