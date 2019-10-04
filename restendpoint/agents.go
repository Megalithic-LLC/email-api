package restendpoint

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/dgrijalva/jwt-go"
	"github.com/docktermj/go-logger/logger"
	"github.com/on-prem-net/email-api/model"
	"github.com/gorilla/context"
	"github.com/gorilla/mux"
)

func (self *RestEndpoint) createAgent(w http.ResponseWriter, req *http.Request) {
	logger.Tracef("RestEndpoint:createAgent")

	// Decode request
	type createAgentRequestType struct {
		Agent model.Agent `json:"agent"`
	}
	var createAgentRequest createAgentRequestType
	if err := json.NewDecoder(req.Body).Decode(&createAgentRequest); err != nil {
		logger.Errorf("Failed decoding agent: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	agentID := createAgentRequest.Agent.ID

	// Validate unclaimed agent
	isMember, err := self.redisClient.SIsMember("unclaimed-agents", agentID).Result()
	if err != nil {
		logger.Errorf("Failed looking up unclaimed agent: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if !isMember {
		logger.Warnf("Attempt to claim non-existent agent %v", agentID)
		w.WriteHeader(http.StatusNotFound)
		return
	}

	// Generate an agent token
	currentUserID := context.Get(req, "currentUserID").(string)
	tokenString, err := self.generateAgentTokenString(currentUserID, agentID)
	if err != nil {
		logger.Errorf("Failed generating token for agent: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Store the token
	redisKey := fmt.Sprintf("tok:%v", tokenString)
	if _, err := self.redisClient.Set(redisKey, "1", 0).Result(); err != nil {
		logger.Errorf("Failed storing agent token: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Communicate the token to the agent, which it will need for future requests
	agentStream := self.agentStreamEndpoint.FindAgentStream(agentID)
	if agentStream == nil {
		if _, err := self.redisClient.Del(redisKey).Result(); err != nil {
			logger.Errorf("Failed deleting agent token: %v", err)
		}
		result := map[string]interface{}{
			"errors": []JsonApiError{
				JsonApiError{
					Status: fmt.Sprintf("%v", http.StatusNotFound),
					Title:  "Agent Connection Failed",
					Detail: "Failed contacting agent; make sure it is running and connected to the internet",
				},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		if err := json.NewEncoder(w).Encode(result); err != nil {
			logger.Errorf("Error encoding response: %v", err)
		}
		return
	}
	if _, err := agentStream.SendClaimRequest(tokenString); err != nil {
		logger.Errorf("Failed claiming agent: %v", err)
		result := map[string]interface{}{
			"errors": []JsonApiError{
				JsonApiError{
					Status: fmt.Sprintf("%v", http.StatusInternalServerError),
					Title:  "Claim Agent Failed",
					Detail: "Failed claiming agent; make sure it is running and connected to the internet",
				},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		if err := json.NewEncoder(w).Encode(result); err != nil {
			logger.Errorf("Error encoding response: %v", err)
		}
		return
	}

	// Create agent
	agent := model.Agent{
		ID:          agentID,
		OwnerUserID: context.Get(req, "currentUserID").(string),
	}
	if err := self.db.Create(&agent).Error; err != nil {
		logger.Errorf("Failed creating new agent: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Remove unclaimed agent reference
	if _, err := self.redisClient.SRem("unclaimed-agents", agentID).Result(); err != nil {
		logger.Errorf("Failed removing unclaimed agent: %v", err)
	}

	// Send result
	result := map[string]interface{}{}
	result["agent"] = agent
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(result); err != nil {
		logger.Errorf("Error encoding response: %v", err)
	}
}

func (self *RestEndpoint) getAgent(w http.ResponseWriter, req *http.Request) {
	id := mux.Vars(req)["id"]
	logger.Tracef("RestEndpoint:getAgent(%s)", id)

	// Find agent
	var agent model.Agent
	searchFor := &model.Agent{ID: id}
	if res := self.db.Where(searchFor).Limit(1).First(&agent); res.RecordNotFound() {
		w.WriteHeader(http.StatusNotFound)
		return
	} else if res.Error != nil {
		logger.Errorf("Failed finding agent: %v", res.Error)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if agent.OwnerUserID != context.Get(req, "currentUserID").(string) {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	// Load related service instances
	if err := self.loadServiceInstances(&agent); err != nil {
		logger.Errorf("Failed loading service instances for agent: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Load related snapshots
	if err := self.loadSnapshots(&agent); err != nil {
		logger.Errorf("Failed loading snapshots for agent: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Send result
	result := map[string]interface{}{}
	result["agent"] = agent
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(result); err != nil {
		logger.Errorf("Error encoding response: %v", err)
	}
}

func (self *RestEndpoint) getAgents(w http.ResponseWriter, req *http.Request) {
	logger.Tracef("RestEndpoint:getAgents")

	// Find agents
	agents := []*model.Agent{}
	searchFor := &model.Agent{OwnerUserID: context.Get(req, "currentUserID").(string)}
	res := self.db.Where(searchFor).Find(&agents)
	if res.Error != nil {
		logger.Errorf("Failed finding all agents: %v", res.Error)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Load related service instances
	for _, agent := range agents {
		if err := self.loadServiceInstances(agent); err != nil {
			logger.Errorf("Failed loading service instances for agent: %v", res.Error)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}

	// Load related snapshots
	for _, agent := range agents {
		if err := self.loadSnapshots(agent); err != nil {
			logger.Errorf("Failed loading snapshots for agent: %v", res.Error)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}

	// Send result
	result := map[string]interface{}{}
	result["agents"] = agents
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(result); err != nil {
		logger.Errorf("Error encoding response: %v", err)
	}
}

func (self *RestEndpoint) loadServiceInstances(agent *model.Agent) error {
	var snapshots []model.ServiceInstance
	searchFor := &model.ServiceInstance{AgentID: agent.ID}
	if err := self.db.Where(searchFor).Find(&snapshots).Error; err != nil {
		logger.Errorf("Failed loading service instances for agent: %v", err)
		return err
	}
	agent.ServiceInstanceIDs = []string{}
	for _, serviceInstance := range snapshots {
		agent.ServiceInstanceIDs = append(agent.ServiceInstanceIDs, serviceInstance.ID)
	}
	return nil
}

func (self *RestEndpoint) loadSnapshots(agent *model.Agent) error {
	var snapshots []model.Snapshot
	searchFor := &model.Snapshot{AgentID: agent.ID}
	if err := self.db.Where(searchFor).Find(&snapshots).Error; err != nil {
		logger.Errorf("Failed loading snapshots for agent: %v", err)
		return err
	}
	agent.SnapshotIDs = []string{}
	for _, snapshot := range snapshots {
		agent.SnapshotIDs = append(agent.SnapshotIDs, snapshot.ID)
	}
	return nil
}

func (self *RestEndpoint) generateAgentTokenString(agentID, userID string) (string, error) {
	claims := &TokenClaims{
		AgentID: agentID,
		UserID:  userID,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: 0,
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecret)
}
