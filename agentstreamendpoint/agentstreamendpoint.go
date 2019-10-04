//go:generate protoc --proto_path=emailproto --go_out=plugins=grpc:emailproto email.proto
package agentstreamendpoint

import (
	"github.com/docktermj/go-logger/logger"
	"github.com/go-redis/redis"
	"github.com/gorilla/websocket"
	"github.com/jinzhu/gorm"
)

type AgentStreamEndpoint struct {
	agentStreamsByAgentID map[string]*AgentStream
	db                    *gorm.DB
	redisClient           *redis.Client
}

func New(db *gorm.DB, redisClient *redis.Client) *AgentStreamEndpoint {
	self := AgentStreamEndpoint{
		agentStreamsByAgentID: map[string]*AgentStream{},
		db:                    db,
		redisClient:           redisClient,
	}
	return &self
}

func (self *AgentStreamEndpoint) FindAgentStream(agentID string) *AgentStream {
	return self.agentStreamsByAgentID[agentID]
}

func (self *AgentStreamEndpoint) NewAgentStream(agentID string, conn *websocket.Conn) *AgentStream {
	logger.Tracef("AgentStreamEndpoint:NewAgentStream(%s)", agentID)
	agentStream := &AgentStream{
		agentID:  agentID,
		conn:     conn,
		endpoint: self,
		pending:  map[uint64]*Call{},
		nextID:   2147483648,
	}
	self.agentStreamsByAgentID[agentID] = agentStream
	return agentStream
}
