package agentstreamendpoint

import (
	"github.com/docktermj/go-logger/logger"
	"github.com/on-prem-net/email-api/agentstreamendpoint/emailproto"
	"github.com/on-prem-net/email-api/model"
)

func (self *AgentStream) handleGetDomainsRequest(requestId uint64, getDomainsReq emailproto.GetDomainsRequest) {
	logger.Tracef("AgentStream:handleGetDomainsRequest(%d)", requestId)

	domains := []model.Domain{}
	searchFor := &model.Domain{AgentID: self.agentID}
	if err := self.endpoint.db.Where(searchFor).Find(&domains).Error; err != nil {
		logger.Errorf("Failed loading domains: %v", err)
		self.SendErrorResponse(requestId, err)
		return
	}

	self.SendGetDomainsResponse(requestId, domains)
}
