package agentstreamendpoint

import (
	"github.com/docktermj/go-logger/logger"
	"github.com/on-prem-net/email-api/agentstreamendpoint/emailproto"
	"github.com/on-prem-net/email-api/model"
)

func (self *AgentStream) handleGetServiceInstancesRequest(requestId uint64, getSnapshotsReq emailproto.GetServiceInstancesRequest) {
	logger.Tracef("AgentStream:handleGetServiceInstancesRequest(%d)", requestId)

	serviceInstances := []model.ServiceInstance{}
	searchFor := &model.ServiceInstance{AgentID: self.agentID}
	if err := self.endpoint.db.Where(searchFor).Find(&serviceInstances).Error; err != nil {
		logger.Errorf("Failed loading service instances: %v", err)
		self.SendErrorResponse(requestId, err)
		return
	}

	self.SendGetServiceInstancesResponse(requestId, serviceInstances)
}
