package agentstreamendpoint

import (
	"github.com/on-prem-net/email-api/agentstreamendpoint/emailproto"
)

func (self *AgentStream) SendGetSnapshotChunksMissingResponse(requestId uint64, chunks []uint32) error {
	res := emailproto.ServerMessage{
		Id: requestId,
		MessageType: &emailproto.ServerMessage_GetSnapshotChunksMissingResponse{
			GetSnapshotChunksMissingResponse: &emailproto.GetSnapshotChunksMissingResponse{
				Chunks: chunks,
			},
		},
	}
	return self.SendResponse(res)
}
