package restendpoint

import (
	"encoding/json"
	"net/http"

	"github.com/Megalithic-LLC/on-prem-email-api/model"
	"github.com/docktermj/go-logger/logger"
	"github.com/gorilla/mux"
)

func (self *RestEndpoint) getPlan(w http.ResponseWriter, req *http.Request) {
	id := mux.Vars(req)["id"]
	logger.Tracef("RestEndpoint:getPlan(%s)", id)

	// Find
	var plan model.Plan
	{
		searchFor := &model.Plan{ID: id}
		if res := self.db.Where(searchFor).Limit(1).First(&plan); res.RecordNotFound() {
			w.WriteHeader(http.StatusNotFound)
			return
		} else if res.Error != nil {
			logger.Errorf("Failed finding plan: %v", res.Error)
			sendInternalServerError(w)
			return
		}
	}

	// Send result
	result := map[string]interface{}{}
	result["plan"] = plan
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(result); err != nil {
		logger.Errorf("Error encoding response: %v", err)
	}
}

func (self *RestEndpoint) getPlans(w http.ResponseWriter, req *http.Request) {
	logger.Tracef("RestEndpoint:getPlans()")

	// Find
	plans := []model.Plan{}
	searchFor := &model.Plan{Visible: true}
	res := self.db.Where(searchFor).Find(&plans)
	if res.Error != nil {
		logger.Errorf("Failed finding all plans: %v", res.Error)
		sendInternalServerError(w)
		return
	}

	// Send result
	result := map[string]interface{}{}
	result["plans"] = plans
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(result); err != nil {
		logger.Errorf("Error encoding response: %v", err)
	}
}
