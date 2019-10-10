package agentstreamendpoint

import (
	"github.com/on-prem-net/email-api/agentstreamendpoint/emailproto"
	"github.com/on-prem-net/email-api/model"
)

func (self *AgentStream) SendGetEndpointsResponse(requestId uint64, endpoints []model.Endpoint) error {
	res := emailproto.ServerMessage{
		Id: requestId,
		MessageType: &emailproto.ServerMessage_GetEndpointsResponse{
			GetEndpointsResponse: &emailproto.GetEndpointsResponse{
				Endpoints: EndpointsAsProtobuf(endpoints),
			},
		},
	}
	return self.SendResponse(res)
}
