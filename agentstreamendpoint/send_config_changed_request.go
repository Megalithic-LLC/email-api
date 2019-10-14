package agentstreamendpoint

import (
	"github.com/docktermj/go-logger/logger"
	"github.com/Megalithic-LLC/on-prem-email-api/agentstreamendpoint/emailproto"
)

func (self *AgentStream) SendConfigChangedRequest() (*emailproto.ClientMessage, error) {
	logger.Tracef("AgentStream:SendConfigChangedRequest()")

	hashesByTable, err := self.calcConfigHashes()
	if err != nil {
		return nil, err
	}

	req := emailproto.ServerMessage{
		MessageType: &emailproto.ServerMessage_ConfigChangedRequest{
			ConfigChangedRequest: &emailproto.ConfigChangedRequest{
				HashesByTable: hashesByTable,
			},
		},
	}
	return self.SendRequest(req)
}
