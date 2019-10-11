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

var (
	validEndpointProtocols = []string{"imap", "lmtp", "smtp", "submission"}
	validEndpointTypes     = []string{"tcp", "unix"}
)

func (self *RestEndpoint) createEndpoint(w http.ResponseWriter, req *http.Request) {
	logger.Tracef("RestEndpoint:createEndpoint()")

	// Decode request
	type createEndpointRequestType struct {
		Endpoint model.Endpoint `json:"endpoint"`
	}
	var createEndpointRequest createEndpointRequestType
	if err := json.NewDecoder(req.Body).Decode(&createEndpointRequest); err != nil {
		logger.Errorf("Failed decoding endpoint: %v", err)
		sendBadRequestError(w, err)
		return
	}
	endpoint := createEndpointRequest.Endpoint

	// Validate
	if validationErrors, err := self.validateEndpoint(&endpoint); err != nil {
		logger.Errorf("Failure validating endpoint: %v", err)
		sendInternalServerError(w)
		return
	} else if len(validationErrors) > 0 {
		sendErrors(w, validationErrors)
		return
	}

	// Store
	if endpoint.ID == "" {
		endpoint.ID = xid.New().String()
	}
	currentUserID := context.Get(req, "currentUserID").(string)
	endpoint.CreatedByUserID = currentUserID
	if err := self.db.Create(&endpoint).Error; err != nil {
		logger.Errorf("Failed creating new endpoint: %v", err)
		sendInternalServerError(w)
		return
	}

	// Notify streaming agent
	if agentStream := self.agentStreamEndpoint.FindAgentStream(endpoint.AgentID); agentStream != nil {
		agentStream.SendConfigChangedRequest()
	}

	// Send result
	result := map[string]interface{}{}
	result["endpoint"] = endpoint
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(result); err != nil {
		logger.Errorf("Error encoding response: %v", err)
	}
}
func (self *RestEndpoint) deleteEndpoint(w http.ResponseWriter, req *http.Request) {
	id := mux.Vars(req)["id"]
	logger.Tracef("RestEndpoint:deleteEndpoint(%s)", id)

	// Find
	var endpoint model.Endpoint
	{
		searchFor := &model.Endpoint{ID: id}
		if res := self.db.Where(searchFor).Limit(1).First(&endpoint); res.RecordNotFound() {
			w.WriteHeader(http.StatusNotFound)
			return
		} else if res.Error != nil {
			logger.Errorf("Failed finding endpoint: %v", res.Error)
			sendInternalServerError(w)
			return
		}
	}

	// Delete
	if err := self.db.Delete(&endpoint).Error; err != nil {
		logger.Errorf("Failed deleting endpoint: %v", err)
		sendInternalServerError(w)
		return
	}

	// Notify streaming agent
	if agentStream := self.agentStreamEndpoint.FindAgentStream(endpoint.AgentID); agentStream != nil {
		agentStream.SendConfigChangedRequest()
	}

	// Send result
	w.WriteHeader(http.StatusNoContent)
}

func (self *RestEndpoint) getEndpoint(w http.ResponseWriter, req *http.Request) {
	id := mux.Vars(req)["id"]
	logger.Tracef("RestEndpoint:getEndpoint(%s)", id)

	// Find
	var endpoint model.Endpoint
	{
		searchFor := &model.Endpoint{ID: id}
		if res := self.db.Where(searchFor).Limit(1).First(&endpoint); res.RecordNotFound() {
			w.WriteHeader(http.StatusNotFound)
			return
		} else if res.Error != nil {
			logger.Errorf("Failed finding endpoint: %v", res.Error)
			sendInternalServerError(w)
			return
		}
	}

	// Send result
	result := map[string]interface{}{}
	result["endpoint"] = endpoint
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(result); err != nil {
		logger.Errorf("Error encoding response: %v", err)
	}
}

func (self *RestEndpoint) getEndpoints(w http.ResponseWriter, req *http.Request) {
	logger.Tracef("RestEndpoint:getEndpoints()")

	// Find
	currentUserID := context.Get(req, "currentUserID").(string)
	endpoints := []model.Endpoint{}
	searchFor := &model.Endpoint{CreatedByUserID: currentUserID}
	res := self.db.Where(searchFor).Find(&endpoints)
	if res.Error != nil {
		logger.Errorf("Failed finding all endpoints: %v", res.Error)
		sendInternalServerError(w)
		return
	}

	// Send result
	result := map[string]interface{}{}
	result["endpoints"] = endpoints
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(result); err != nil {
		logger.Errorf("Error encoding response: %v", err)
	}
}

func (self *RestEndpoint) validateEndpoint(endpoint *model.Endpoint) ([]JsonApiError, error) {
	errs := []JsonApiError{}
	if endpoint.AgentID == "" {
		err := JsonApiError{
			Status: fmt.Sprintf("%d", http.StatusBadRequest),
			Title:  "Validation Error",
			Detail: "An agent id is required",
		}
		errs = append(errs, err)
	}
	protocolIsValid := false
	for _, protocol := range validEndpointProtocols {
		if protocol == endpoint.Protocol {
			protocolIsValid = true
			break
		}
	}
	if !protocolIsValid {
		err := JsonApiError{
			Status: fmt.Sprintf("%d", http.StatusBadRequest),
			Title:  "Validation Error",
			Detail: fmt.Sprintf("A valid protocol is required; use one of %v", validEndpointProtocols),
		}
		errs = append(errs, err)
	}
	typeIsValid := false
	for _, endpointType := range validEndpointTypes {
		if endpointType == endpoint.Type {
			typeIsValid = true
			break
		}
	}
	if !typeIsValid {
		err := JsonApiError{
			Status: fmt.Sprintf("%d", http.StatusBadRequest),
			Title:  "Validation Error",
			Detail: fmt.Sprintf("A valid type is required; use one of %v", validEndpointTypes),
		}
		errs = append(errs, err)
	}
	return errs, nil
}
