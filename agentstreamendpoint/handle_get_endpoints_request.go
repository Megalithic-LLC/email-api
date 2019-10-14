package agentstreamendpoint

import (
	"github.com/docktermj/go-logger/logger"
	"github.com/Megalithic-LLC/on-prem-email-api/agentstreamendpoint/emailproto"
	"github.com/Megalithic-LLC/on-prem-email-api/model"
)

func (self *AgentStream) handleGetEndpointsRequest(requestId uint64, getEndpointsReq emailproto.GetEndpointsRequest) {
	logger.Tracef("AgentStream:handleGetEndpointsRequest(%d)", requestId)

	endpoints := []model.Endpoint{}
	searchFor := &model.Endpoint{AgentID: self.agentID}
	if err := self.endpoint.db.Where(searchFor).Find(&endpoints).Error; err != nil {
		logger.Errorf("Failed loading endpoints: %v", err)
		self.SendErrorResponse(requestId, err)
		return
	}

	self.SendGetEndpointsResponse(requestId, endpoints)
}
