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

func (self *RestEndpoint) createDomain(w http.ResponseWriter, req *http.Request) {
	logger.Tracef("RestEndpoint:createDomain()")

	// Decode request
	type createDomainRequestType struct {
		Domain model.Domain `json:"domain"`
	}
	var createDomainRequest createDomainRequestType
	if err := json.NewDecoder(req.Body).Decode(&createDomainRequest); err != nil {
		logger.Errorf("Failed decoding domain: %v", err)
		sendBadRequestError(w, err)
		return
	}
	domain := createDomainRequest.Domain

	// Validate
	if validationErrors, err := self.validateDomain(&domain); err != nil {
		logger.Errorf("Failure validating domain: %v", err)
		sendInternalServerError(w)
		return
	} else if len(validationErrors) > 0 {
		sendErrors(w, validationErrors)
		return
	}

	// Store
	if domain.ID == "" {
		domain.ID = xid.New().String()
	}
	currentUserID := context.Get(req, "currentUserID").(string)
	domain.CreatedByUserID = currentUserID
	if err := self.db.Create(&domain).Error; err != nil {
		logger.Errorf("Failed creating new domain: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Notify streaming agent
	if agentStream := self.agentStreamEndpoint.FindAgentStream(domain.AgentID); agentStream != nil {
		agentStream.SendConfigChangedRequest()
	}

	// Send result
	result := map[string]interface{}{}
	result["domain"] = domain
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(result); err != nil {
		logger.Errorf("Error encoding response: %v", err)
	}
}
func (self *RestEndpoint) deleteDomain(w http.ResponseWriter, req *http.Request) {
	id := mux.Vars(req)["id"]
	logger.Tracef("RestEndpoint:deleteDomain(%s)", id)

	// Find
	var domain model.Domain
	{
		searchFor := &model.Domain{ID: id}
		if res := self.db.Where(searchFor).Limit(1).First(&domain); res.RecordNotFound() {
			w.WriteHeader(http.StatusNotFound)
			return
		} else if res.Error != nil {
			logger.Errorf("Failed finding domain: %v", res.Error)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}

	// Delete
	if err := self.db.Delete(&domain).Error; err != nil {
		logger.Errorf("Failed deleting domain: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Notify streaming agent
	if agentStream := self.agentStreamEndpoint.FindAgentStream(domain.AgentID); agentStream != nil {
		agentStream.SendConfigChangedRequest()
	}

	// Send result
	w.WriteHeader(http.StatusNoContent)
}

func (self *RestEndpoint) getDomain(w http.ResponseWriter, req *http.Request) {
	id := mux.Vars(req)["id"]
	logger.Tracef("RestEndpoint:getDomain(%s)", id)

	// Find
	var domain model.Domain
	{
		searchFor := &model.Domain{ID: id}
		if res := self.db.Where(searchFor).Limit(1).First(&domain); res.RecordNotFound() {
			w.WriteHeader(http.StatusNotFound)
			return
		} else if res.Error != nil {
			logger.Errorf("Failed finding domain: %v", res.Error)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}

	// Send result
	result := map[string]interface{}{}
	result["domain"] = domain
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(result); err != nil {
		logger.Errorf("Error encoding response: %v", err)
	}
}

func (self *RestEndpoint) getDomains(w http.ResponseWriter, req *http.Request) {
	logger.Tracef("RestEndpoint:getDomains()")

	// Find
	currentUserID := context.Get(req, "currentUserID").(string)
	domains := []model.Domain{}
	searchFor := &model.Domain{CreatedByUserID: currentUserID}
	res := self.db.Where(searchFor).Find(&domains)
	if res.Error != nil {
		logger.Errorf("Failed finding all domains: %v", res.Error)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Send result
	result := map[string]interface{}{}
	result["domains"] = domains
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(result); err != nil {
		logger.Errorf("Error encoding response: %v", err)
	}
}

func (self *RestEndpoint) validateDomain(domain *model.Domain) ([]JsonApiError, error) {
	errs := []JsonApiError{}
	if domain.AgentID == "" {
		err := JsonApiError{
			Status: fmt.Sprintf("%d", http.StatusBadRequest),
			Title:  "Validation Error",
			Detail: "An agent id is required",
		}
		errs = append(errs, err)
	}
	if domain.ServiceInstanceID == "" {
		err := JsonApiError{
			Status: fmt.Sprintf("%d", http.StatusBadRequest),
			Title:  "Validation Error",
			Detail: "A service instance id is required",
		}
		errs = append(errs, err)
	}
	if domain.Name == "" {
		err := JsonApiError{
			Status: fmt.Sprintf("%d", http.StatusBadRequest),
			Title:  "Validation Error",
			Detail: "A name is required",
		}
		errs = append(errs, err)
	}
	return errs, nil
}
