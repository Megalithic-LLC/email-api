package restendpoint

import (
	"errors"
	"net/http"

	"github.com/docktermj/go-logger/logger"
	"github.com/gorilla/websocket"
)

var agentStreamUpgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func (self *RestEndpoint) agentStream(w http.ResponseWriter, req *http.Request) {
	logger.Tracef("RestEndpoint:agentStream")

	var agentID string
	if values := req.Header["X-Agentid"]; len(values) != 1 {
		logger.Errorf("Agent failed to present an id: %+v", req)
		sendBadRequestError(w, errors.New("Agent failed to present an X-Agentid header"))
		return
	} else {
		agentID = values[0]
	}

	conn, err := agentStreamUpgrader.Upgrade(w, req, nil)
	if err != nil {
		logger.Errorf("Failed upgrading to agentstream websocket: %v", err)
		return
	}
	defer conn.Close()

	logger.Infof("Agent %s is now streaming", agentID)
	agentStream := self.agentStreamEndpoint.NewAgentStream(agentID, conn)
	agentStream.Reader()
}
