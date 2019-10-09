package agentstreamendpoint

import (
	"fmt"

	"github.com/docktermj/go-logger/logger"
	"github.com/on-prem-net/email-api/agentstreamendpoint/emailproto"
	"github.com/on-prem-net/email-api/model"
)

func (self *AgentStream) handleSetSnapshotChunkRequest(requestId uint64, setSnapshotChunkReq emailproto.SetSnapshotChunkRequest) {
	logger.Tracef("AgentStream:handleSetSnapshotChunkRequest(%d)", requestId)

	snapshot := model.Snapshot{}
	searchFor := &model.Snapshot{AgentID: self.agentID, ID: setSnapshotChunkReq.SnapshotId}
	if err := self.endpoint.db.Where(searchFor).Limit(1).First(&snapshot).Error; err != nil {
		logger.Errorf("Failed loading snapshot: %v", err)
		self.SendErrorResponse(requestId, err)
		return
	}

	snapshotChunk := model.SnapshotChunk{
		SnapshotID: snapshot.ID,
		Number:     setSnapshotChunkReq.Number,
		Data:       setSnapshotChunkReq.Data,
	}
	if err := self.endpoint.db.Save(&snapshotChunk).Error; err != nil {
		logger.Errorf("Failed updating snapshot chunk: %v", err)
		self.SendErrorResponse(requestId, err)
		return
	}

	key := fmt.Sprintf("snapshot/%s/needed-chunks", snapshot.ID)
	if err := self.endpoint.redisClient.SRem(key, snapshotChunk.Number).Err(); err != nil {
		logger.Errorf("Failed removing needed snapshot chunk from redis: %v", err)
		self.SendErrorResponse(requestId, err)
		return
	}

	self.SendAckResponse(requestId)
}
