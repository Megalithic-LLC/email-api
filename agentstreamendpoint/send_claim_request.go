package agentstreamendpoint

import (
	"github.com/on-prem-net/email-api/agentstreamendpoint/emailproto"
	"github.com/docktermj/go-logger/logger"
)

func (self *AgentStream) SendClaimRequest(token string) (*emailproto.ClientMessage, error) {
	logger.Tracef("AgentStream:SendClaimRequest(%s)", token)
	req := emailproto.ServerMessage{
		MessageType: &emailproto.ServerMessage_ClaimRequest{
			ClaimRequest: &emailproto.ClaimRequest{
				Token: token,
			},
		},
	}
	return self.SendRequest(req)
}
