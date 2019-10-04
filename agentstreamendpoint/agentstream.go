package agentstreamendpoint

import (
	"errors"
	"sync"
	"time"

	"github.com/on-prem-net/email-api/agentstreamendpoint/emailproto"
	"github.com/docktermj/go-logger/logger"
	"github.com/golang/protobuf/proto"
	"github.com/gorilla/websocket"
)

type AgentStream struct {
	agentID  string
	conn     *websocket.Conn
	endpoint *AgentStreamEndpoint
	mutex    sync.Mutex
	nextID   uint64
	pending  map[uint64]*Call
}

func (self *AgentStream) getNextID() uint64 {
	id := self.nextID
	self.nextID++
	return id
}

func (self *AgentStream) Reader() {
	var err error
	for err == nil {

		// Read message
		mt, rawMessage, err := self.conn.ReadMessage()
		if err != nil {
			logger.Errorf("Failed reading message: %v", err)
			break
		}

		if mt != websocket.BinaryMessage {
			continue
		}

		// Decode message
		var message emailproto.ClientMessage
		if err := proto.Unmarshal(rawMessage, &message); err != nil {
			logger.Errorf("Failed decoding message: %v", err)
			break
		}

		self.mutex.Lock()
		call, isResponse := self.pending[message.Id]
		if isResponse {
			delete(self.pending, message.Id)
			self.mutex.Unlock()
			call.Res = &message
			call.Done <- true
		} else {
			self.mutex.Unlock()
			self.route(message)
		}

	}

	// Terminate all calls
	self.mutex.Lock()
	for _, call := range self.pending {
		call.Error = err
		call.Done <- true
	}
	self.mutex.Unlock()
}

func (self *AgentStream) SendRequest(req emailproto.ServerMessage) (*emailproto.ClientMessage, error) {
	self.mutex.Lock()
	req.Id = self.getNextID()
	call := NewCall(req)
	self.pending[req.Id] = call

	// Encode request
	rawMessage, err := proto.Marshal(&req)
	if err != nil {
		logger.Errorf("Failed encoding request: %v", err)
		defer self.mutex.Unlock()
		return nil, err
	}

	// Send request
	if err := self.conn.WriteMessage(websocket.BinaryMessage, rawMessage); err != nil {
		delete(self.pending, req.Id)
		defer self.mutex.Unlock()
		return nil, err
	}

	self.mutex.Unlock()

	select {
	case <-call.Done:
	case <-time.After(2 * time.Second):
		call.Error = errors.New("request timeout")
	}

	return call.Res, call.Error
}

func (self *AgentStream) SendResponse(res emailproto.ServerMessage) error {
	// Encode response
	rawMessage, err := proto.Marshal(&res)
	if err != nil {
		logger.Errorf("Failed encoding response: %v", err)
		return err
	}

	// Send response
	return self.conn.WriteMessage(websocket.BinaryMessage, rawMessage)
}
