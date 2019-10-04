package agentstreamendpoint

import (
	"github.com/on-prem-net/email-api/agentstreamendpoint/emailproto"
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
