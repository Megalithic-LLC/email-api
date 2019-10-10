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

func (self *RestEndpoint) createServiceInstance(w http.ResponseWriter, req *http.Request) {
	logger.Tracef("RestEndpoint:createServiceInstance()")

	// Decode request
	type createServiceInstanceRequestType struct {
		ServiceInstance model.ServiceInstance `json:"serviceInstance"`
	}
	var createServiceInstanceRequest createServiceInstanceRequestType
	if err := json.NewDecoder(req.Body).Decode(&createServiceInstanceRequest); err != nil {
		logger.Errorf("Failed decoding service instance: %v", err)
		sendBadRequestError(w, err)
		return
	}
	serviceInstance := createServiceInstanceRequest.ServiceInstance

	// Validate
	if validationErrors, err := self.validateServiceInstance(&serviceInstance); err != nil {
		logger.Errorf("Failure validating service instance: %v", err)
		sendInternalServerError(w)
		return
	} else if len(validationErrors) > 0 {
		sendErrors(w, validationErrors)
		return
	}

	// Set defaults
	if serviceInstance.ID == "" {
		serviceInstance.ID = xid.New().String()
	}
	currentUserID := context.Get(req, "currentUserID").(string)
	serviceInstance.CreatedByUserID = currentUserID

	imapEndpoint := model.Endpoint{
		ID:                xid.New().String(),
		AgentID:           serviceInstance.AgentID,
		ServiceInstanceID: serviceInstance.ID,
		Protocol:          "imap",
		Type:              "tcp",
		Port:              8143,
		Enabled:           true,
		CreatedByUserID:   currentUserID,
	}

	lmtpEndpoint := model.Endpoint{
		ID:                xid.New().String(),
		AgentID:           serviceInstance.AgentID,
		ServiceInstanceID: serviceInstance.ID,
		Protocol:          "lmtp",
		Type:              "unix",
		Enabled:           true,
		CreatedByUserID:   currentUserID,
	}

	smtpEndpoint := model.Endpoint{
		ID:                xid.New().String(),
		AgentID:           serviceInstance.AgentID,
		ServiceInstanceID: serviceInstance.ID,
		Protocol:          "smtp",
		Type:              "tcp",
		Port:              8025,
		Enabled:           true,
		CreatedByUserID:   currentUserID,
	}

	submissionEndpoint := model.Endpoint{
		ID:                xid.New().String(),
		AgentID:           serviceInstance.AgentID,
		ServiceInstanceID: serviceInstance.ID,
		Protocol:          "submission",
		Type:              "tcp",
		Port:              8587,
		Enabled:           true,
		CreatedByUserID:   currentUserID,
	}

	// Store
	tx := self.db.Begin()
	defer tx.Rollback()
	if err := tx.Create(&serviceInstance).Error; err != nil {
		logger.Errorf("Failed creating service instance: %v", err)
		tx.Rollback()
		sendInternalServerError(w)
		return
	}
	if err := tx.Create(&imapEndpoint).Error; err != nil {
		logger.Errorf("Failed creating default imap endpoint for new service instance: %v", err)
		tx.Rollback()
		sendInternalServerError(w)
		return
	}
	if err := tx.Create(&lmtpEndpoint).Error; err != nil {
		logger.Errorf("Failed creating default lmtp endpoint for new service instance: %v", err)
		tx.Rollback()
		sendInternalServerError(w)
		return
	}
	if err := tx.Create(&smtpEndpoint).Error; err != nil {
		logger.Errorf("Failed creating default smtp endpoint for new service instance: %v", err)
		tx.Rollback()
		sendInternalServerError(w)
		return
	}
	if err := tx.Create(&submissionEndpoint).Error; err != nil {
		logger.Errorf("Failed creating default smtp submission endpoint for new service instance: %v", err)
		tx.Rollback()
		sendInternalServerError(w)
		return
	}
	if err := tx.Commit().Error; err != nil {
		logger.Errorf("Failed creating new service instance: %v", err)
		tx.Rollback()
		sendInternalServerError(w)
		return
	}

	serviceInstance.EndpointIDs = []string{
		imapEndpoint.ID,
		lmtpEndpoint.ID,
		smtpEndpoint.ID,
		submissionEndpoint.ID,
	}

	// Send result
	result := map[string]interface{}{}
	result["serviceInstance"] = serviceInstance
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(result); err != nil {
		logger.Errorf("Error encoding response: %v", err)
	}
}

func (self *RestEndpoint) deleteServiceInstance(w http.ResponseWriter, req *http.Request) {
	id := mux.Vars(req)["id"]
	logger.Tracef("RestEndpoint:deleteServiceInstance(%s)", id)

	// Find
	var serviceInstance model.ServiceInstance
	{
		searchFor := &model.ServiceInstance{ID: id}
		if res := self.db.Where(searchFor).Limit(1).First(&serviceInstance); res.RecordNotFound() {
			w.WriteHeader(http.StatusNotFound)
			return
		} else if res.Error != nil {
			logger.Errorf("Failed finding service instance: %v", res.Error)
			sendInternalServerError(w)
			return
		}
	}

	// Delete
	tx := self.db.Begin()
	defer tx.Rollback()
	if err := tx.Delete(&model.Account{ServiceInstanceID: id}).Error; err != nil {
		logger.Errorf("Failed deleting accounts related to service instance: %v", err)
		tx.Rollback()
		sendInternalServerError(w)
		return
	}
	if err := tx.Delete(&model.Domain{ServiceInstanceID: id}).Error; err != nil {
		logger.Errorf("Failed deleting domains related to service instance: %v", err)
		tx.Rollback()
		sendInternalServerError(w)
		return
	}
	if err := tx.Delete(&model.Endpoint{ServiceInstanceID: id}).Error; err != nil {
		logger.Errorf("Failed deleting endpoints related to service instance: %v", err)
		tx.Rollback()
		sendInternalServerError(w)
		return
	}
	if err := tx.Delete(&serviceInstance).Error; err != nil {
		logger.Errorf("Failed deleting service instance: %v", err)
		tx.Rollback()
		sendInternalServerError(w)
		return
	}
	if err := tx.Commit().Error; err != nil {
		logger.Errorf("Failed deleting service instance: %v", err)
		tx.Rollback()
		sendInternalServerError(w)
		return
	}

	// Notify streaming agent
	if agentStream := self.agentStreamEndpoint.FindAgentStream(serviceInstance.AgentID); agentStream != nil {
		agentStream.SendConfigChangedRequest()
	}

	// Send result
	w.WriteHeader(http.StatusNoContent)
}

func (self *RestEndpoint) getServiceInstance(w http.ResponseWriter, req *http.Request) {
	id := mux.Vars(req)["id"]
	logger.Tracef("RestEndpoint:getServiceInstance(%s)", id)

	// Find
	var serviceInstance model.ServiceInstance
	{
		currentUserID := context.Get(req, "currentUserID").(string)
		searchFor := &model.ServiceInstance{ID: id, CreatedByUserID: currentUserID}
		if res := self.db.Where(searchFor).Limit(1).First(&serviceInstance); res.RecordNotFound() {
			w.WriteHeader(http.StatusNotFound)
			return
		} else if res.Error != nil {
			logger.Errorf("Failed finding service instance: %v", res.Error)
			sendInternalServerError(w)
			return
		}
	}

	// Load related accounts
	if err := self.loadAccounts(&serviceInstance); err != nil {
		logger.Errorf("Failed loading accounts for service instance: %v", err)
		sendInternalServerError(w)
		return
	}

	// Load related domains
	if err := self.loadDomains(&serviceInstance); err != nil {
		logger.Errorf("Failed loading domains for service instance: %v", err)
		sendInternalServerError(w)
		return
	}

	// Load related endpoints
	if err := self.loadEndpoints(&serviceInstance); err != nil {
		logger.Errorf("Failed loading endpoints for service instance: %v", err)
		sendInternalServerError(w)
		return
	}

	// Send result
	result := map[string]interface{}{}
	result["serviceInstance"] = serviceInstance
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(result); err != nil {
		logger.Errorf("Error encoding response: %v", err)
	}
}

func (self *RestEndpoint) getServiceInstances(w http.ResponseWriter, req *http.Request) {
	logger.Tracef("RestEndpoint:getServiceInstances()")

	// Find
	serviceInstances := []*model.ServiceInstance{}
	searchFor := &model.ServiceInstance{}
	res := self.db.Where(searchFor).Find(&serviceInstances)
	if res.Error != nil {
		logger.Errorf("Failed finding all serviceInstances: %v", res.Error)
		sendInternalServerError(w)
		return
	}

	// Load related foreign keys
	for _, serviceInstance := range serviceInstances {
		// Load related accounts
		if err := self.loadAccounts(serviceInstance); err != nil {
			logger.Errorf("Failed loading accounts for service instance: %v", res.Error)
			sendInternalServerError(w)
			return
		}

		// Load related domains
		if err := self.loadDomains(serviceInstance); err != nil {
			logger.Errorf("Failed loading domains for service instance: %v", res.Error)
			sendInternalServerError(w)
			return
		}

		// Load related endpoints
		if err := self.loadEndpoints(serviceInstance); err != nil {
			logger.Errorf("Failed loading endpoints for service instance: %v", res.Error)
			sendInternalServerError(w)
			return
		}
	}

	// Send result
	result := map[string]interface{}{}
	result["serviceInstances"] = serviceInstances
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(result); err != nil {
		logger.Errorf("Error encoding response: %v", err)
	}
}

func (self *RestEndpoint) loadAccounts(serviceInstance *model.ServiceInstance) error {
	var accounts []model.Account
	searchFor := &model.Account{ServiceInstanceID: serviceInstance.ID}
	if err := self.db.Where(searchFor).Find(&accounts).Error; err != nil {
		logger.Errorf("Failed loading accounts for service instance: %v", err)
		return err
	}
	serviceInstance.AccountIDs = []string{}
	for _, account := range accounts {
		serviceInstance.AccountIDs = append(serviceInstance.AccountIDs, account.ID)
	}
	return nil
}

func (self *RestEndpoint) loadDomains(serviceInstance *model.ServiceInstance) error {
	var domains []model.Domain
	searchFor := &model.Domain{ServiceInstanceID: serviceInstance.ID}
	if err := self.db.Where(searchFor).Find(&domains).Error; err != nil {
		logger.Errorf("Failed loading domains for service instance: %v", err)
		return err
	}
	serviceInstance.DomainIDs = []string{}
	for _, domain := range domains {
		serviceInstance.DomainIDs = append(serviceInstance.DomainIDs, domain.ID)
	}
	return nil
}

func (self *RestEndpoint) loadEndpoints(serviceInstance *model.ServiceInstance) error {
	var endpoints []model.Endpoint
	searchFor := &model.Endpoint{ServiceInstanceID: serviceInstance.ID}
	if err := self.db.Where(searchFor).Find(&endpoints).Error; err != nil {
		logger.Errorf("Failed loading endpoints for service instance: %v", err)
		return err
	}
	serviceInstance.EndpointIDs = []string{}
	for _, endpoint := range endpoints {
		serviceInstance.EndpointIDs = append(serviceInstance.EndpointIDs, endpoint.ID)
	}
	return nil
}

func (self *RestEndpoint) validateServiceInstance(serviceInstance *model.ServiceInstance) ([]JsonApiError, error) {
	errs := []JsonApiError{}
	if serviceInstance.AgentID == "" {
		err := JsonApiError{
			Status: fmt.Sprintf("%d", http.StatusBadRequest),
			Title:  "Validation Error",
			Detail: "An agent reference is required",
		}
		errs = append(errs, err)
	}
	if serviceInstance.ServiceID == "" {
		err := JsonApiError{
			Status: fmt.Sprintf("%d", http.StatusBadRequest),
			Title:  "Validation Error",
			Detail: "A service reference is required",
		}
		errs = append(errs, err)
	}
	if serviceInstance.PlanID == "" {
		err := JsonApiError{
			Status: fmt.Sprintf("%d", http.StatusBadRequest),
			Title:  "Validation Error",
			Detail: "A plan reference is required",
		}
		errs = append(errs, err)
	}

	// Only one service instance of a given service can be activated within an agent
	{
		var serviceInstance model.ServiceInstance
		searchFor := &model.ServiceInstance{AgentID: serviceInstance.AgentID, ServiceID: serviceInstance.ServiceID}
		if res := self.db.Where(searchFor).Limit(1).First(&serviceInstance); res.Error == nil {
			err := JsonApiError{
				Status: fmt.Sprintf("%d", http.StatusInternalServerError),
				Title:  "Validation Error",
				Detail: "A service may not be added to an agent more than once",
			}
			errs = append(errs, err)
		} else if res.Error != nil && !res.RecordNotFound() {
			logger.Errorf("Failed looking for existing service instance: %v", res.Error)
			err := JsonApiError{
				Status: fmt.Sprintf("%d", http.StatusInternalServerError),
				Title:  "Validation Error",
				Detail: "An internal server error has occurred",
			}
			errs = append(errs, err)
		}
	}

	return errs, nil
}
