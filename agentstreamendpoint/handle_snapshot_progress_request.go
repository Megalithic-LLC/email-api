package agentstreamendpoint

import (
	"github.com/docktermj/go-logger/logger"
	"github.com/on-prem-net/email-api/agentstreamendpoint/emailproto"
	"github.com/on-prem-net/email-api/model"
)

func (self *AgentStream) handleSnapshotProgressRequest(requestId uint64, snapshotProgressReq emailproto.SnapshotProgressRequest) {
	logger.Tracef("AgentStream:handleSnapshotProgressRequest(%d)", requestId)

	snapshot := model.Snapshot{}
	searchFor := &model.Snapshot{AgentID: self.agentID, ID: snapshotProgressReq.SnapshotId}
	if err := self.endpoint.db.Where(searchFor).Limit(1).First(&snapshot).Error; err != nil {
		logger.Errorf("Failed loading snapshot: %v", err)
		self.SendErrorResponse(requestId, err)
		return
	}

	updates := model.Snapshot{
		Progress: snapshotProgressReq.Progress,
		Size:     snapshotProgressReq.Size,
	}
	if err := self.endpoint.db.Model(&snapshot).Updates(&updates).Error; err != nil {
		logger.Errorf("Failed updating snapshot with progress: %v", err)
		self.SendErrorResponse(requestId, err)
		return
	}

	self.SendAckResponse(requestId)
}
