package restendpoint

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/docktermj/go-logger/logger"
	"github.com/gorilla/mux"
	"github.com/on-prem-net/email-api/model"
)

type serviceResType struct {
	model.Service
	PlanIDs []string `json:"plans"`
}

func (self *RestEndpoint) getService(w http.ResponseWriter, req *http.Request) {
	id := mux.Vars(req)["id"]
	logger.Tracef("RestEndpoint:getService(%s)", id)

	// Find
	var service model.Service
	{
		searchFor := &model.Service{ID: id}
		if res := self.db.Where(searchFor).Limit(1).First(&service); res.RecordNotFound() {
			w.WriteHeader(http.StatusNotFound)
			return
		} else if res.Error != nil {
			logger.Errorf("Failed finding service: %v", res.Error)
			sendInternalServerError(w)
			return
		}
	}

	// Load referenced plans
	serviceRes, err := self.loadServicePlans(service)
	if err != nil {
		logger.Errorf("Failed loading plans for service: %v", err)
		sendInternalServerError(w)
		return
	}

	// Send result
	result := map[string]interface{}{}
	result["service"] = serviceRes
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(result); err != nil {
		logger.Errorf("Error encoding response: %v", err)
	}
}

func (self *RestEndpoint) getServices(w http.ResponseWriter, req *http.Request) {
	logger.Tracef("RestEndpoint:getServices(%v)", req.URL.Query())

	// Find
	services := []model.Service{}
	searchFor := parseServicesFilter(req)
	res := self.db.Where(searchFor).Find(&services)
	if res.Error != nil {
		logger.Errorf("Failed finding all services: %v", res.Error)
		sendInternalServerError(w)
		return
	}

	// Load referenced plans
	servicesRes := []*serviceResType{}
	for _, service := range services {
		serviceRes, err := self.loadServicePlans(service)
		if err != nil {
			logger.Errorf("Failed loading plans for service: %v", res.Error)
			sendInternalServerError(w)
			return
		}
		servicesRes = append(servicesRes, serviceRes)
	}

	// Send result
	result := map[string]interface{}{}
	result["services"] = servicesRes
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(result); err != nil {
		logger.Errorf("Error encoding response: %v", err)
	}
}

func (self *RestEndpoint) loadServicePlans(service model.Service) (*serviceResType, error) {
	var plans []model.Plan
	searchFor := &model.Plan{ServiceID: service.ID}
	if err := self.db.Where(searchFor).Find(&plans).Error; err != nil {
		logger.Errorf("Failed loading plans for service: %v", err)
		return nil, err
	}
	serviceRes := serviceResType{
		Service: service,
		PlanIDs: []string{},
	}
	for _, plan := range plans {
		serviceRes.PlanIDs = append(serviceRes.PlanIDs, plan.ID)
	}
	return &serviceRes, nil
}

func parseServicesFilter(req *http.Request) *model.Service {
	query := req.URL.Query()
	searchFor := &model.Service{}
	if visible, ok := query["filter[visible]"]; ok {
		if fmt.Sprintf("%v", visible) == "[true]" {
			searchFor.Visible = true
		}
	}
	return searchFor
}
