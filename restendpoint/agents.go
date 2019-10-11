package restendpoint

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/dgrijalva/jwt-go"
	"github.com/docktermj/go-logger/logger"
	"github.com/gorilla/context"
	"github.com/gorilla/mux"
	"github.com/on-prem-net/email-api/model"
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
		sendBadRequestError(w, err)
		return
	}
	agentID := createAgentRequest.Agent.ID

	// Validate unclaimed agent
	isMember, err := self.redisClient.SIsMember("unclaimed-agents", agentID).Result()
	if err != nil {
		logger.Errorf("Failed looking up unclaimed agent: %v", err)
		sendInternalServerError(w)
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
		sendInternalServerError(w)
		return
	}

	// Store the token
	redisKey := fmt.Sprintf("tok:%v", tokenString)
	if _, err := self.redisClient.Set(redisKey, "1", 0).Result(); err != nil {
		logger.Errorf("Failed storing agent token: %v", err)
		sendInternalServerError(w)
		return
	}

	// Communicate the token to the agent, which it will need for future requests
	agentStream := self.agentStreamEndpoint.FindAgentStream(agentID)
	if agentStream == nil {
		if _, err := self.redisClient.Del(redisKey).Result(); err != nil {
			logger.Errorf("Failed deleting agent token: %v", err)
		}
		sendErrors(w, []JsonApiError{
			JsonApiError{
				Status: fmt.Sprintf("%v", http.StatusNotFound),
				Title:  "Agent Connection Failed",
				Detail: "Failed contacting agent; make sure it is running and connected to the internet",
			},
		})
		return
	}
	if _, err := agentStream.SendClaimRequest(tokenString); err != nil {
		logger.Errorf("Failed claiming agent: %v", err)
		sendErrors(w, []JsonApiError{
			JsonApiError{
				Status: fmt.Sprintf("%v", http.StatusInternalServerError),
				Title:  "Claim Agent Failed",
				Detail: "Failed claiming agent; make sure it is running and connected to the internet",
			},
		})
		return
	}

	// Create agent
	agent := model.Agent{
		ID:              agentID,
		CreatedByUserID: currentUserID,
	}
	if err := self.db.Create(&agent).Error; err != nil {
		logger.Errorf("Failed creating new agent: %v", err)
		sendInternalServerError(w)
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
		sendInternalServerError(w)
		return
	}
	if agent.CreatedByUserID != context.Get(req, "currentUserID").(string) {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	// Load related accounts
	if err := self.loadAccounts(&agent); err != nil {
		logger.Errorf("Failed loading accounts for agent: %v", err)
		sendInternalServerError(w)
		return
	}

	// Load related domains
	if err := self.loadDomains(&agent); err != nil {
		logger.Errorf("Failed loading domains for agent: %v", err)
		sendInternalServerError(w)
		return
	}

	// Load related endpoints
	if err := self.loadEndpoints(&agent); err != nil {
		logger.Errorf("Failed loading endpoints for agent: %v", err)
		sendInternalServerError(w)
		return
	}

	// Load related snapshots
	if err := self.loadSnapshots(&agent); err != nil {
		logger.Errorf("Failed loading snapshots for agent: %v", err)
		sendInternalServerError(w)
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
	searchFor := &model.Agent{CreatedByUserID: context.Get(req, "currentUserID").(string)}
	res := self.db.Where(searchFor).Find(&agents)
	if res.Error != nil {
		logger.Errorf("Failed finding all agents: %v", res.Error)
		sendInternalServerError(w)
		return
	}

	// Load related accounts
	for _, agent := range agents {
		if err := self.loadAccounts(agent); err != nil {
			logger.Errorf("Failed loading accounts for agent: %v", res.Error)
			sendInternalServerError(w)
			return
		}
	}

	// Load related domains
	for _, agent := range agents {
		if err := self.loadDomains(agent); err != nil {
			logger.Errorf("Failed loading domains for agent: %v", res.Error)
			sendInternalServerError(w)
			return
		}
	}

	// Load related endpoints
	for _, agent := range agents {
		if err := self.loadEndpoints(agent); err != nil {
			logger.Errorf("Failed loading endpoints for agent: %v", res.Error)
			sendInternalServerError(w)
			return
		}
	}

	// Load related snapshots
	for _, agent := range agents {
		if err := self.loadSnapshots(agent); err != nil {
			logger.Errorf("Failed loading snapshots for agent: %v", res.Error)
			sendInternalServerError(w)
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

func (self *RestEndpoint) loadAccounts(agent *model.Agent) error {
	var accounts []model.Account
	searchFor := &model.Account{AgentID: agent.ID}
	if err := self.db.Where(searchFor).Find(&accounts).Error; err != nil {
		logger.Errorf("Failed loading accounts for service instance: %v", err)
		return err
	}
	agent.AccountIDs = []string{}
	for _, account := range accounts {
		agent.AccountIDs = append(agent.AccountIDs, account.ID)
	}
	return nil
}

func (self *RestEndpoint) loadDomains(agent *model.Agent) error {
	var domains []model.Domain
	searchFor := &model.Domain{AgentID: agent.ID}
	if err := self.db.Where(searchFor).Find(&domains).Error; err != nil {
		logger.Errorf("Failed loading domains for service instance: %v", err)
		return err
	}
	agent.DomainIDs = []string{}
	for _, domain := range domains {
		agent.DomainIDs = append(agent.DomainIDs, domain.ID)
	}
	return nil
}

func (self *RestEndpoint) loadEndpoints(agent *model.Agent) error {
	var endpoints []model.Endpoint
	searchFor := &model.Endpoint{AgentID: agent.ID}
	if err := self.db.Where(searchFor).Find(&endpoints).Error; err != nil {
		logger.Errorf("Failed loading endpoints for service instance: %v", err)
		return err
	}
	agent.EndpointIDs = []string{}
	for _, endpoint := range endpoints {
		agent.EndpointIDs = append(agent.EndpointIDs, endpoint.ID)
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
