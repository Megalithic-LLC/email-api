package agentstreamendpoint

import (
	"github.com/Megalithic-LLC/on-prem-email-api/agentstreamendpoint/emailproto"
	"github.com/Megalithic-LLC/on-prem-email-api/model"
)

func (self *AgentStream) SendGetDomainsResponse(requestId uint64, domains []model.Domain) error {
	res := emailproto.ServerMessage{
		Id: requestId,
		MessageType: &emailproto.ServerMessage_GetDomainsResponse{
			GetDomainsResponse: &emailproto.GetDomainsResponse{
				Domains: DomainsAsProtobuf(domains),
			},
		},
	}
	return self.SendResponse(res)
}
