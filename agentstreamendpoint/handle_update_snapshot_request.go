package agentstreamendpoint

import (
	"fmt"
	"math"

	"github.com/docktermj/go-logger/logger"
	"github.com/Megalithic-LLC/on-prem-email-api/agentstreamendpoint/emailproto"
	"github.com/Megalithic-LLC/on-prem-email-api/model"
)

const (
	chunkSize = uint32(1000000)
)

func (self *AgentStream) handleUpdateSnapshotRequest(requestId uint64, updateSnapshotReq emailproto.UpdateSnapshotRequest) {
	logger.Tracef("AgentStream:handleUpdateSnapshotRequest(%d)", requestId)

	snapshot := model.Snapshot{}
	searchFor := &model.Snapshot{AgentID: self.agentID, ID: updateSnapshotReq.Snapshot.Id}
	if err := self.endpoint.db.Where(searchFor).Limit(1).First(&snapshot).Error; err != nil {
		logger.Errorf("Failed loading snapshot: %v", err)
		self.SendErrorResponse(requestId, err)
		return
	}

	prevSize := snapshot.Size

	updates := SnapshotFromProtobuf(updateSnapshotReq.Snapshot)

	if err := self.endpoint.db.Model(&snapshot).Updates(updates).Error; err != nil {
		logger.Errorf("Failed updating snapshot : %v", err)
		self.SendErrorResponse(requestId, err)
		return
	}

	if updates.Size > prevSize {
		chunkCount := uint32(math.Ceil(float64(updates.Size) / float64(chunkSize)))
		logger.Debugf("Chunks: %d size:%d", chunkCount, updates.Size)
		chunkNumbers := []string{}
		for i := uint32(0); i < chunkCount; i++ {
			chunkNumbers = append(chunkNumbers, fmt.Sprintf("%d", i))
		}
		key := fmt.Sprintf("snapshot/%s/needed-chunks", snapshot.ID)
		if err := self.endpoint.redisClient.SAdd(key, chunkNumbers).Err(); err != nil {
			logger.Errorf("Failed writing snapshot chunks required into redis: %v", err)
		}
	}

	self.SendAckResponse(requestId)
}
