package agentstreamendpoint

import (
	"github.com/on-prem-net/email-api/agentstreamendpoint/emailproto"
)

func (self *AgentStream) SendAckResponse(requestId uint64) error {
	ackRes := emailproto.ServerMessage{
		Id: requestId,
		MessageType: &emailproto.ServerMessage_AckResponse{
			AckResponse: &emailproto.AckResponse{},
		},
	}
	return self.SendResponse(ackRes)
}
