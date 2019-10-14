package agentstreamendpoint

import (
	"github.com/Megalithic-LLC/on-prem-email-api/agentstreamendpoint/emailproto"
)

func (self *AgentStream) SendStartupResponse(requestId uint64) error {

	hashesByTable, err := self.calcConfigHashes()
	if err != nil {
		return err
	}

	res := emailproto.ServerMessage{
		Id: requestId,
		MessageType: &emailproto.ServerMessage_StartupResponse{
			StartupResponse: &emailproto.StartupResponse{
				ConfigHashesByTable: hashesByTable,
			},
		},
	}
	return self.SendResponse(res)
}
