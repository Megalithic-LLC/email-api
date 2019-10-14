package agentstreamendpoint

import (
	"github.com/Megalithic-LLC/on-prem-email-api/agentstreamendpoint/emailproto"
	"github.com/Megalithic-LLC/on-prem-email-api/model"
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
