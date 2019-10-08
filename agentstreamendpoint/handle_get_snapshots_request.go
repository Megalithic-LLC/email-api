package agentstreamendpoint

import (
	"github.com/docktermj/go-logger/logger"
	"github.com/on-prem-net/email-api/agentstreamendpoint/emailproto"
	"github.com/on-prem-net/email-api/model"
)

func (self *AgentStream) handleGetSnapshotsRequest(requestId uint64, getSnapshotsReq emailproto.GetSnapshotsRequest) {
	logger.Tracef("AgentStream:handleGetSnapshotsRequest(%d)", requestId)

	snapshots := []model.Snapshot{}
	searchFor := &model.Snapshot{AgentID: self.agentID}
	if err := self.endpoint.db.Where(searchFor).Find(&snapshots).Error; err != nil {
		logger.Errorf("Failed loading snapshots: %v", err)
		self.SendErrorResponse(requestId, err)
		return
	}

	self.SendGetSnapshotsResponse(requestId, snapshots)
}
