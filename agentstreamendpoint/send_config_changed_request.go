package agentstreamendpoint

import (
	"github.com/on-prem-net/email-api/agentstreamendpoint/emailproto"
	"github.com/docktermj/go-logger/logger"
)

func (self *AgentStream) SendConfigChangedRequest() (*emailproto.ClientMessage, error) {
	logger.Tracef("AgentStream:SendConfigChangedRequest()")

	hashesByTable := map[string][]byte{}
	// TODO

	req := emailproto.ServerMessage{
		MessageType: &emailproto.ServerMessage_ConfigChangedRequest{
			ConfigChangedRequest: &emailproto.ConfigChangedRequest{
				HashesByTable: hashesByTable,
			},
		},
	}
	return self.SendRequest(req)
}
