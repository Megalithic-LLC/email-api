package restendpoint

import (
	"encoding/json"
	"net/http"

	"github.com/docktermj/go-logger/logger"
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
	if serviceInstance.AgentID == "" || serviceInstance.ServiceID == "" || serviceInstance.PlanID == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if serviceInstance.ID == "" {
		serviceInstance.ID = xid.New().String()
	}

	// Store
	if err := self.db.Create(&serviceInstance).Error; err != nil {
		logger.Errorf("Failed creating new service instance: %v", err)
		sendInternalServerError(w)
		return
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

func (self *RestEndpoint) getServiceInstance(w http.ResponseWriter, req *http.Request) {
	id := mux.Vars(req)["id"]
	logger.Tracef("RestEndpoint:getServiceInstance(%s)", id)

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

	// Load related accounts
	if err := self.loadAccounts(&serviceInstance); err != nil {
		logger.Errorf("Failed loading accounts for service instance: %v", err)
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

	// Load related service instances
	for _, serviceInstance := range serviceInstances {
		if err := self.loadAccounts(serviceInstance); err != nil {
			logger.Errorf("Failed loading accounts for service instance: %v", res.Error)
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
