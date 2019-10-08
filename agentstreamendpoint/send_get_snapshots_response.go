package agentstreamendpoint

import (
	"github.com/on-prem-net/email-api/agentstreamendpoint/emailproto"
	"github.com/on-prem-net/email-api/model"
)

func (self *AgentStream) SendGetSnapshotsResponse(requestId uint64, snapshots []model.Snapshot) error {
	getSnapshotsRes := emailproto.ServerMessage{
		Id: requestId,
		MessageType: &emailproto.ServerMessage_GetSnapshotsResponse{
			GetSnapshotsResponse: &emailproto.GetSnapshotsResponse{
				Snapshots: SnapshotsAsProtobuf(snapshots),
			},
		},
	}
	return self.SendResponse(getSnapshotsRes)
}
