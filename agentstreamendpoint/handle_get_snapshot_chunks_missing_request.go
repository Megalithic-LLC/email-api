package agentstreamendpoint

import (
	"fmt"
	"strconv"

	"github.com/docktermj/go-logger/logger"
	"github.com/on-prem-net/email-api/agentstreamendpoint/emailproto"
	"github.com/on-prem-net/email-api/model"
)

func (self *AgentStream) handleGetSnapshotChunksMissingRequest(requestId uint64, getSnapshotChunksMissingReq emailproto.GetSnapshotChunksMissingRequest) {
	logger.Tracef("AgentStream:handleGetSnapshotChunksMissingRequest(%d)", requestId)

	snapshot := model.Snapshot{}
	searchFor := &model.Snapshot{AgentID: self.agentID, ID: getSnapshotChunksMissingReq.SnapshotId}
	if err := self.endpoint.db.Where(searchFor).Limit(1).First(&snapshot).Error; err != nil {
		logger.Errorf("Failed loading snapshot: %v", err)
		self.SendErrorResponse(requestId, err)
		return
	}

	if snapshot.Size == 0 {
		self.SendGetSnapshotChunksMissingResponse(requestId, []uint32{})
		return
	}

	key := fmt.Sprintf("snapshot/%s/needed-chunks", snapshot.ID)
	chunkNumbers, err := self.endpoint.redisClient.SMembers(key).Result()
	if err != nil {
		logger.Errorf("Failed reading needed snapshot chunks from redis: %v", err)
		self.SendErrorResponse(requestId, err)
		return
	}

	nums := []uint32{}
	for _, chunkNumber := range chunkNumbers {
		if num, err := strconv.ParseUint(chunkNumber, 10, 32); err == nil {
			nums = append(nums, uint32(num))
		}
	}

	self.SendGetSnapshotChunksMissingResponse(requestId, nums)
}
