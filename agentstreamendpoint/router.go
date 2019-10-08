package agentstreamendpoint

import (
	"github.com/docktermj/go-logger/logger"
	"github.com/on-prem-net/email-api/agentstreamendpoint/emailproto"
)

func (self *AgentStream) route(message emailproto.ClientMessage) {
	logger.Tracef("AgentStream:route()")

	switch message.MessageType.(type) {

	case *emailproto.ClientMessage_GetAccountsRequest:
		getAccountsRequest := message.GetGetAccountsRequest()
		self.handleGetAccountsRequest(message.Id, *getAccountsRequest)

	case *emailproto.ClientMessage_StartupRequest:
		startupRequest := message.GetStartupRequest()
		self.handleStartupRequest(message.Id, *startupRequest)

	}
}
