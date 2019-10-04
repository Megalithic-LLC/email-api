package restendpoint

import (
	"encoding/json"
	"net/http"

	"github.com/docktermj/go-logger/logger"
	"github.com/Megalithic-LLC/on-prem-admin-api/model"
	"github.com/gorilla/mux"
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
		w.WriteHeader(http.StatusBadRequest)
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
		w.WriteHeader(http.StatusInternalServerError)
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
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
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
	serviceInstances := []model.ServiceInstance{}
	searchFor := &model.ServiceInstance{}
	res := self.db.Where(searchFor).Find(&serviceInstances)
	if res.Error != nil {
		logger.Errorf("Failed finding all serviceInstances: %v", res.Error)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Send result
	result := map[string]interface{}{}
	result["serviceInstances"] = serviceInstances
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(result); err != nil {
		logger.Errorf("Error encoding response: %v", err)
	}
}
