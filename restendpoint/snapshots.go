package restendpoint

import (
	"encoding/json"
	"net/http"

	"github.com/on-prem-net/email-api/model"
	"github.com/docktermj/go-logger/logger"
	"github.com/gorilla/mux"
	"github.com/rs/xid"
)

func (self *RestEndpoint) createSnapshot(w http.ResponseWriter, req *http.Request) {
	logger.Tracef("RestEndpoint:createSnapshot()")

	// Decode request
	type createSnapshotRequestType struct {
		Snapshot model.Snapshot `json:"snapshot"`
	}
	var createSnapshotRequest createSnapshotRequestType
	if err := json.NewDecoder(req.Body).Decode(&createSnapshotRequest); err != nil {
		logger.Errorf("Failed decoding snapshot: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	snapshot := createSnapshotRequest.Snapshot

	// Validate
	if snapshot.AgentID == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if snapshot.ID == "" {
		snapshot.ID = xid.New().String()
	}

	// Store
	if err := self.db.Create(&snapshot).Error; err != nil {
		logger.Errorf("Failed creating new snapshot: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Notify streaming agent
	if agentStream := self.agentStreamEndpoint.FindAgentStream(snapshot.AgentID); agentStream != nil {
		agentStream.SendConfigChangedRequest()
	}

	// Send result
	result := map[string]interface{}{}
	result["snapshot"] = snapshot
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(result); err != nil {
		logger.Errorf("Error encoding response: %v", err)
	}
}

func (self *RestEndpoint) getSnapshot(w http.ResponseWriter, req *http.Request) {
	id := mux.Vars(req)["id"]
	logger.Tracef("RestEndpoint:getSnapshot(%s)", id)

	// Find
	var snapshot model.Snapshot
	{
		searchFor := &model.Snapshot{ID: id}
		if res := self.db.Where(searchFor).Limit(1).First(&snapshot); res.RecordNotFound() {
			w.WriteHeader(http.StatusNotFound)
			return
		} else if res.Error != nil {
			logger.Errorf("Failed finding snapshot: %v", res.Error)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}

	// Send result
	result := map[string]interface{}{}
	result["snapshot"] = snapshot
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(result); err != nil {
		logger.Errorf("Error encoding response: %v", err)
	}
}

func (self *RestEndpoint) getSnapshots(w http.ResponseWriter, req *http.Request) {
	logger.Tracef("RestEndpoint:getSnapshots()")

	// Find
	snapshots := []model.Snapshot{}
	searchFor := &model.Snapshot{}
	res := self.db.Where(searchFor).Find(&snapshots)
	if res.Error != nil {
		logger.Errorf("Failed finding all snapshots: %v", res.Error)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Send result
	result := map[string]interface{}{}
	result["snapshots"] = snapshots
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(result); err != nil {
		logger.Errorf("Error encoding response: %v", err)
	}
}
