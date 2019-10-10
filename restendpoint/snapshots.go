package restendpoint

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/docktermj/go-logger/logger"
	"github.com/gorilla/context"
	"github.com/gorilla/mux"
	"github.com/on-prem-net/email-api/model"
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
		sendBadRequestError(w, err)
		return
	}
	snapshot := createSnapshotRequest.Snapshot

	// Validate
	if validationErrors, err := self.validateSnapshot(&snapshot); err != nil {
		logger.Errorf("Failure validating snapshot: %v", err)
		sendInternalServerError(w)
		return
	} else if len(validationErrors) > 0 {
		sendErrors(w, validationErrors)
		return
	}

	// Store
	if snapshot.ID == "" {
		snapshot.ID = xid.New().String()
	}
	currentUserID := context.Get(req, "currentUserID").(string)
	snapshot.CreatedByUserID = currentUserID
	if err := self.db.Create(&snapshot).Error; err != nil {
		logger.Errorf("Failed creating new snapshot: %v", err)
		sendInternalServerError(w)
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

func (self *RestEndpoint) deleteSnapshot(w http.ResponseWriter, req *http.Request) {
	id := mux.Vars(req)["id"]
	logger.Tracef("RestEndpoint:deleteSnapshot(%s)", id)

	// Find
	var snapshot model.Snapshot
	{
		searchFor := &model.Snapshot{ID: id}
		if res := self.db.Where(searchFor).Limit(1).First(&snapshot); res.RecordNotFound() {
			w.WriteHeader(http.StatusNotFound)
			return
		} else if res.Error != nil {
			logger.Errorf("Failed finding snapshot: %v", res.Error)
			sendInternalServerError(w)
			return
		}
	}

	// Delete
	if err := self.db.Delete(&snapshot).Error; err != nil {
		logger.Errorf("Failed deleting snapshot: %v", err)
		sendInternalServerError(w)
		return
	}

	// Notify streaming agent
	if agentStream := self.agentStreamEndpoint.FindAgentStream(snapshot.AgentID); agentStream != nil {
		agentStream.SendConfigChangedRequest()
	}

	// Send result
	w.WriteHeader(http.StatusNoContent)
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
			sendInternalServerError(w)
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
	currentUserID := context.Get(req, "currentUserID").(string)
	snapshots := []model.Snapshot{}
	searchFor := &model.Snapshot{CreatedByUserID: currentUserID}
	res := self.db.Where(searchFor).Find(&snapshots)
	if res.Error != nil {
		logger.Errorf("Failed finding all snapshots: %v", res.Error)
		sendInternalServerError(w)
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

func (self *RestEndpoint) validateSnapshot(snapshot *model.Snapshot) ([]JsonApiError, error) {
	errs := []JsonApiError{}
	if snapshot.AgentID == "" {
		err := JsonApiError{
			Status: fmt.Sprintf("%d", http.StatusBadRequest),
			Title:  "Validation Error",
			Detail: "An agent id is required",
		}
		errs = append(errs, err)
	}
	return errs, nil
}
