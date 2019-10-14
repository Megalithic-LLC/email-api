package agentstreamendpoint

import (
	"errors"

	"github.com/docktermj/go-logger/logger"
	"github.com/Megalithic-LLC/on-prem-email-api/agentstreamendpoint/emailproto"
	"github.com/Megalithic-LLC/on-prem-email-api/model"
)

func (self *AgentStream) handleStartupRequest(requestId uint64, startupReq emailproto.StartupRequest) {
	logger.Tracef("AgentStream:handleStartupRequest(%d)", requestId)

	// This API service is specific to the email agent
	if startupReq.ServiceId != "blmkmfd5jj89vu275l5g" {
		logger.Errorf("Agent is trying to connect for some other service %s", startupReq.ServiceId)
		self.SendErrorResponse(requestId, errors.New("This API service only handles email agents"))
	}

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

	self.SendStartupResponse(requestId)
}
