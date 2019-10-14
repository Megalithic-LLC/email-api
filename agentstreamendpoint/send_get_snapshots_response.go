package agentstreamendpoint

import (
	"github.com/Megalithic-LLC/on-prem-email-api/agentstreamendpoint/emailproto"
	"github.com/Megalithic-LLC/on-prem-email-api/model"
)

func (self *AgentStream) SendGetSnapshotsResponse(requestId uint64, snapshots []model.Snapshot) error {
	res := emailproto.ServerMessage{
		Id: requestId,
		MessageType: &emailproto.ServerMessage_GetSnapshotsResponse{
			GetSnapshotsResponse: &emailproto.GetSnapshotsResponse{
				Snapshots: SnapshotsAsProtobuf(snapshots),
			},
		},
	}
	return self.SendResponse(res)
}
