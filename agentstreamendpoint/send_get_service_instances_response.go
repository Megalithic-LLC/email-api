package agentstreamendpoint

import (
	"github.com/on-prem-net/email-api/agentstreamendpoint/emailproto"
	"github.com/on-prem-net/email-api/model"
)

func (self *AgentStream) SendGetServiceInstancesResponse(requestId uint64, serviceInstances []model.ServiceInstance) error {
	res := emailproto.ServerMessage{
		Id: requestId,
		MessageType: &emailproto.ServerMessage_GetServiceInstancesResponse{
			GetServiceInstancesResponse: &emailproto.GetServiceInstancesResponse{
				ServiceInstances: ServiceInstancesAsProtobuf(serviceInstances),
			},
		},
	}
	return self.SendResponse(res)
}
