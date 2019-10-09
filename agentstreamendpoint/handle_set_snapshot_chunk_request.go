package agentstreamendpoint

import (
	"fmt"
	"math"

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

	// Update progress. How many chunks remaining?
	chunkNumbers, err := self.endpoint.redisClient.SMembers(key).Result()
	if err != nil {
		logger.Errorf("Failed reading needed snapshot chunks from redis: %v", err)
		self.SendErrorResponse(requestId, err)
		return
	}
	chunksRemainingCount := len(chunkNumbers)
	totalChunksCount := int(math.Ceil(float64(snapshot.Size) / float64(chunkSize)))
	chunksCompletedCount := totalChunksCount - chunksRemainingCount
	progress := 100 * float32(chunksCompletedCount) / float32(totalChunksCount)

	// Update progress from 0..100% to 50..100% (transfer is 2nd half of range, file snapshot is 1st)
	progress = 50 + progress/2

	updates := model.Snapshot{Progress: progress}
	if err := self.endpoint.db.Model(&snapshot).Updates(&updates).Error; err != nil {
		logger.Errorf("Failed updating snapshot with progress: %v", err)
		self.SendErrorResponse(requestId, err)
		return
	}

	self.SendAckResponse(requestId)
}
