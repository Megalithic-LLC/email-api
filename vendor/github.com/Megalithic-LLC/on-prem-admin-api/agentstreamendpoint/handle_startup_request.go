package agentstreamendpoint

import (
	"github.com/Megalithic-LLC/on-prem-admin-api/agentstreamendpoint/emailproto"
	"github.com/Megalithic-LLC/on-prem-admin-api/model"
	"github.com/docktermj/go-logger/logger"
)

func (self *AgentStream) handleStartupRequest(requestId uint64, startupReq emailproto.StartupRequest) {
	logger.Tracef("AgentStream:handleStartupRequest(%d)", requestId)

	// See if this agent is known
	var agent model.Agent
	searchFor := &model.Agent{ID: self.agentID}
	res := self.endpoint.db.Where(searchFor).First(&agent)
	if res.Error != nil && !res.RecordNotFound() {
		logger.Errorf("Failed looking up agent: %v", res.Error)
		self.SendErrorResponse(requestId, res.Error)
		return
	}
	agentIsKnown := !res.RecordNotFound()

	// If this agent is not yet known, register it as unclaimed
	if !agentIsKnown {
		if err := self.endpoint.redisClient.SAdd("unclaimed-agents", self.agentID).Err(); err != nil {
			logger.Errorf("Failed storing unclaimed agent: %v", err)
			self.SendErrorResponse(requestId, err)
			return
		}
		logger.Infof("Stored unclaimed agent")

	}

	startupRes := emailproto.ServerMessage{
		Id: requestId,
		MessageType: &emailproto.ServerMessage_AckResponse{
			AckResponse: &emailproto.AckResponse{},
		},
	}
	self.SendResponse(startupRes)
}
