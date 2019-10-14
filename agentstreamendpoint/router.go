package agentstreamendpoint

import (
	"github.com/docktermj/go-logger/logger"
	"github.com/Megalithic-LLC/on-prem-email-api/agentstreamendpoint/emailproto"
)

func (self *AgentStream) route(message emailproto.ClientMessage) {
	logger.Tracef("AgentStream:route()")

	switch message.MessageType.(type) {

	case *emailproto.ClientMessage_GetAccountsRequest:
		getAccountsRequest := message.GetGetAccountsRequest()
		self.handleGetAccountsRequest(message.Id, *getAccountsRequest)

	case *emailproto.ClientMessage_GetDomainsRequest:
		getDomainsRequest := message.GetGetDomainsRequest()
		self.handleGetDomainsRequest(message.Id, *getDomainsRequest)

	case *emailproto.ClientMessage_GetEndpointsRequest:
		getEndpointsRequest := message.GetGetEndpointsRequest()
		self.handleGetEndpointsRequest(message.Id, *getEndpointsRequest)

	case *emailproto.ClientMessage_GetSnapshotChunksMissingRequest:
		getSnapshotChunksMissingRequest := message.GetGetSnapshotChunksMissingRequest()
		self.handleGetSnapshotChunksMissingRequest(message.Id, *getSnapshotChunksMissingRequest)

	case *emailproto.ClientMessage_GetSnapshotsRequest:
		getSnapshotsRequest := message.GetGetSnapshotsRequest()
		self.handleGetSnapshotsRequest(message.Id, *getSnapshotsRequest)

	case *emailproto.ClientMessage_SetSnapshotChunkRequest:
		setSnapshotChunkRequest := message.GetSetSnapshotChunkRequest()
		self.handleSetSnapshotChunkRequest(message.Id, *setSnapshotChunkRequest)

	case *emailproto.ClientMessage_StartupRequest:
		startupRequest := message.GetStartupRequest()
		self.handleStartupRequest(message.Id, *startupRequest)

	case *emailproto.ClientMessage_UpdateSnapshotRequest:
		updateSnapshotRequest := message.GetUpdateSnapshotRequest()
		self.handleUpdateSnapshotRequest(message.Id, *updateSnapshotRequest)

	}
}
