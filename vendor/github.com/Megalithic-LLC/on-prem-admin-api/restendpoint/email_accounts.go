package restendpoint

import (
	"encoding/json"
	"net/http"

	"github.com/Megalithic-LLC/on-prem-admin-api/model"
	"github.com/docktermj/go-logger/logger"
	"github.com/gorilla/mux"
	"github.com/rs/xid"
)

func (self *RestEndpoint) createEmailAccount(w http.ResponseWriter, req *http.Request) {
	logger.Tracef("RestEndpoint:createEmailAccount()")

	// Decode request
	type createEmailAccountRequestType struct {
		EmailAccount model.EmailAccount `json:"emailAccount"`
	}
	var createEmailAccountRequest createEmailAccountRequestType
	if err := json.NewDecoder(req.Body).Decode(&createEmailAccountRequest); err != nil {
		logger.Errorf("Failed decoding emailAccount: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	emailAccount := createEmailAccountRequest.EmailAccount

	// Validate
	if emailAccount.AgentID == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if emailAccount.ID == "" {
		emailAccount.ID = xid.New().String()
	}

	// Store
	if err := self.db.Create(&emailAccount).Error; err != nil {
		logger.Errorf("Failed creating new email account: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Notify streaming agent
	if agentStream := self.agentStreamEndpoint.FindAgentStream(emailAccount.AgentID); agentStream != nil {
		agentStream.SendConfigChangedRequest()
	}

	// Send result
	result := map[string]interface{}{}
	result["emailAccount"] = emailAccount
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(result); err != nil {
		logger.Errorf("Error encoding response: %v", err)
	}
}
func (self *RestEndpoint) deleteEmailAccount(w http.ResponseWriter, req *http.Request) {
	id := mux.Vars(req)["id"]
	logger.Tracef("RestEndpoint:deleteEmailAccount(%s)", id)

	// Find
	var emailAccount model.EmailAccount
	{
		searchFor := &model.EmailAccount{ID: id}
		if res := self.db.Where(searchFor).Limit(1).First(&emailAccount); res.RecordNotFound() {
			w.WriteHeader(http.StatusNotFound)
			return
		} else if res.Error != nil {
			logger.Errorf("Failed finding email account: %v", res.Error)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}

	// Delete
	if err := self.db.Delete(&emailAccount).Error; err != nil {
		logger.Errorf("Failed deleting email account: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Notify streaming agent
	if agentStream := self.agentStreamEndpoint.FindAgentStream(emailAccount.AgentID); agentStream != nil {
		agentStream.SendConfigChangedRequest()
	}

	// Send result
	w.WriteHeader(http.StatusNoContent)
}

func (self *RestEndpoint) getEmailAccount(w http.ResponseWriter, req *http.Request) {
	id := mux.Vars(req)["id"]
	logger.Tracef("RestEndpoint:getEmailAccount(%s)", id)

	// Find
	var emailAccount model.EmailAccount
	{
		searchFor := &model.EmailAccount{ID: id}
		if res := self.db.Where(searchFor).Limit(1).First(&emailAccount); res.RecordNotFound() {
			w.WriteHeader(http.StatusNotFound)
			return
		} else if res.Error != nil {
			logger.Errorf("Failed finding emailAccount: %v", res.Error)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}

	// Send result
	result := map[string]interface{}{}
	result["emailAccount"] = emailAccount
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(result); err != nil {
		logger.Errorf("Error encoding response: %v", err)
	}
}

func (self *RestEndpoint) getEmailAccounts(w http.ResponseWriter, req *http.Request) {
	logger.Tracef("RestEndpoint:getEmailAccounts()")

	// Find
	emailAccounts := []model.EmailAccount{}
	searchFor := &model.EmailAccount{}
	res := self.db.Where(searchFor).Find(&emailAccounts)
	if res.Error != nil {
		logger.Errorf("Failed finding all email accounts: %v", res.Error)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Send result
	result := map[string]interface{}{}
	result["emailAccounts"] = emailAccounts
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(result); err != nil {
		logger.Errorf("Error encoding response: %v", err)
	}
}
