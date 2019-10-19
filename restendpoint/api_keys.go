package restendpoint

import (
	"encoding/json"
	"net/http"

	"github.com/Megalithic-LLC/on-prem-email-api/model"
	"github.com/docktermj/go-logger/logger"
	"github.com/gorilla/context"
	"github.com/gorilla/mux"
	"github.com/rs/xid"
)

func (self *RestEndpoint) createApiKey(w http.ResponseWriter, req *http.Request) {
	logger.Tracef("RestEndpoint:createApiKey()")

	currentUserID := context.Get(req, "currentUserID").(string)

	// Decode request
	type createApiKeyRequestType struct {
		ApiKey model.ApiKey `json:"apiKey"`
	}
	var createApiKeyRequest createApiKeyRequestType
	if err := json.NewDecoder(req.Body).Decode(&createApiKeyRequest); err != nil {
		logger.Errorf("Failed decoding API Key: %v", err)
		sendBadRequestError(w, err)
		return
	}
	apiKey := createApiKeyRequest.ApiKey

	// Validate
	if validationErrors, err := self.validateApiKey(&apiKey); err != nil {
		logger.Errorf("Failure validating API Key: %v", err)
		sendInternalServerError(w)
		return
	} else if len(validationErrors) > 0 {
		sendErrors(w, validationErrors)
		return
	}

	// Generate a key if not provided by client
	if apiKey.Key == "" {
		apiKey.Key = xid.New().String()
	}

	// Store
	if apiKey.ID == "" {
		apiKey.ID = xid.New().String()
	}
	apiKey.CreatedByUserID = currentUserID
	if err := self.db.Create(&apiKey).Error; err != nil {
		logger.Errorf("Failed creating new API Key: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Send result
	result := map[string]interface{}{}
	result["apiKey"] = apiKey
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(result); err != nil {
		logger.Errorf("Error encoding response: %v", err)
	}
}
func (self *RestEndpoint) deleteApiKey(w http.ResponseWriter, req *http.Request) {
	id := mux.Vars(req)["id"]
	logger.Tracef("RestEndpoint:deleteApiKey(%s)", id)

	// Find
	var apiKey model.ApiKey
	{
		searchFor := &model.ApiKey{ID: id}
		if res := self.db.Where(searchFor).Limit(1).First(&apiKey); res.RecordNotFound() {
			w.WriteHeader(http.StatusNotFound)
			return
		} else if res.Error != nil {
			logger.Errorf("Failed finding API Key: %v", res.Error)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}

	// Delete
	if err := self.db.Delete(&apiKey).Error; err != nil {
		logger.Errorf("Failed deleting API Key: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Send result
	w.WriteHeader(http.StatusNoContent)
}

func (self *RestEndpoint) getApiKey(w http.ResponseWriter, req *http.Request) {
	id := mux.Vars(req)["id"]
	logger.Tracef("RestEndpoint:getApiKey(%s)", id)

	// Find
	var apiKey model.ApiKey
	{
		searchFor := &model.ApiKey{ID: id}
		if res := self.db.Where(searchFor).Limit(1).First(&apiKey); res.RecordNotFound() {
			w.WriteHeader(http.StatusNotFound)
			return
		} else if res.Error != nil {
			logger.Errorf("Failed finding API Key: %v", res.Error)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}

	// Send result
	result := map[string]interface{}{}
	result["apiKey"] = apiKey
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(result); err != nil {
		logger.Errorf("Error encoding response: %v", err)
	}
}

func (self *RestEndpoint) getApiKeys(w http.ResponseWriter, req *http.Request) {
	logger.Tracef("RestEndpoint:getApiKeys()")

	// Find
	currentUserID := context.Get(req, "currentUserID").(string)
	apiKeys := []model.ApiKey{}
	searchFor := &model.ApiKey{CreatedByUserID: currentUserID}
	res := self.db.Where(searchFor).Find(&apiKeys)
	if res.Error != nil {
		logger.Errorf("Failed finding all API Keys: %v", res.Error)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Send result
	result := map[string]interface{}{}
	result["apiKeys"] = apiKeys
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(result); err != nil {
		logger.Errorf("Error encoding response: %v", err)
	}
}

func (self *RestEndpoint) validateApiKey(apiKey *model.ApiKey) ([]JsonApiError, error) {
	errs := []JsonApiError{}
	return errs, nil
}
