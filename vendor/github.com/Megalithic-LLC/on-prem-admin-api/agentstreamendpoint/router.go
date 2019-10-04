package agentstreamendpoint

import (
	"github.com/Megalithic-LLC/on-prem-admin-api/agentstreamendpoint/emailproto"
	"github.com/docktermj/go-logger/logger"
)

func (self *AgentStream) route(message emailproto.ClientMessage) {
	logger.Tracef("AgentStream:route()")

	switch message.MessageType.(type) {

	case *emailproto.ClientMessage_StartupRequest:
		startupRequest := message.GetStartupRequest()
		self.handleStartupRequest(message.Id, *startupRequest)
	}
}
