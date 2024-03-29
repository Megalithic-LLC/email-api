package agentstreamendpoint

import (
	"github.com/Megalithic-LLC/on-prem-email-api/agentstreamendpoint/emailproto"
)

func (self *AgentStream) SendErrorResponse(requestId uint64, err error) error {
	errorRes := emailproto.ServerMessage{
		Id: requestId,
		MessageType: &emailproto.ServerMessage_ErrorResponse{
			ErrorResponse: &emailproto.ErrorResponse{
				Error: err.Error(),
			},
		},
	}
	return self.SendResponse(errorRes)
}
