package restendpoint

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/mail"
	"strings"

	"github.com/docktermj/go-logger/logger"
	"github.com/gorilla/context"
	"github.com/gorilla/mux"
	"github.com/on-prem-net/email-api/model"
	"github.com/rs/xid"
)

func (self *RestEndpoint) createAccount(w http.ResponseWriter, req *http.Request) {
	logger.Tracef("RestEndpoint:createAccount()")

	// Decode request
	type createAccountRequestType struct {
		Account model.Account `json:"account"`
	}
	var createAccountRequest createAccountRequestType
	if err := json.NewDecoder(req.Body).Decode(&createAccountRequest); err != nil {
		logger.Errorf("Failed decoding account: %v", err)
		sendBadRequestError(w, err)
		return
	}
	account := createAccountRequest.Account

	// Validate
	if validationErrors, err := self.validateAccount(&account); err != nil {
		logger.Errorf("Failure validating account: %v", err)
		sendInternalServerError(w)
		return
	} else if len(validationErrors) > 0 {
		sendErrors(w, validationErrors)
		return
	}

	// Store
	if account.ID == "" {
		account.ID = xid.New().String()
	}
	currentUserID := context.Get(req, "currentUserID").(string)
	account.CreatedByUserID = currentUserID
	if err := self.db.Create(&account).Error; err != nil {
		logger.Errorf("Failed creating new account: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Notify streaming agent
	if agentStream := self.agentStreamEndpoint.FindAgentStream(account.AgentID); agentStream != nil {
		agentStream.SendConfigChangedRequest()
	}

	// Send result
	result := map[string]interface{}{}
	result["account"] = account
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(result); err != nil {
		logger.Errorf("Error encoding response: %v", err)
	}
}
func (self *RestEndpoint) deleteAccount(w http.ResponseWriter, req *http.Request) {
	id := mux.Vars(req)["id"]
	logger.Tracef("RestEndpoint:deleteAccount(%s)", id)

	// Find
	var account model.Account
	{
		searchFor := &model.Account{ID: id}
		if res := self.db.Where(searchFor).Limit(1).First(&account); res.RecordNotFound() {
			w.WriteHeader(http.StatusNotFound)
			return
		} else if res.Error != nil {
			logger.Errorf("Failed finding email account: %v", res.Error)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}

	// Delete
	if err := self.db.Delete(&account).Error; err != nil {
		logger.Errorf("Failed deleting email account: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Notify streaming agent
	if agentStream := self.agentStreamEndpoint.FindAgentStream(account.AgentID); agentStream != nil {
		agentStream.SendConfigChangedRequest()
	}

	// Send result
	w.WriteHeader(http.StatusNoContent)
}

func (self *RestEndpoint) getAccount(w http.ResponseWriter, req *http.Request) {
	id := mux.Vars(req)["id"]
	logger.Tracef("RestEndpoint:getAccount(%s)", id)

	// Find
	var account model.Account
	{
		searchFor := &model.Account{ID: id}
		if res := self.db.Where(searchFor).Limit(1).First(&account); res.RecordNotFound() {
			w.WriteHeader(http.StatusNotFound)
			return
		} else if res.Error != nil {
			logger.Errorf("Failed finding account: %v", res.Error)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}

	// Send result
	result := map[string]interface{}{}
	result["account"] = account
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(result); err != nil {
		logger.Errorf("Error encoding response: %v", err)
	}
}

func (self *RestEndpoint) getAccounts(w http.ResponseWriter, req *http.Request) {
	logger.Tracef("RestEndpoint:getAccounts()")

	// Find
	currentUserID := context.Get(req, "currentUserID").(string)
	accounts := []model.Account{}
	searchFor := &model.Account{CreatedByUserID: currentUserID}
	res := self.db.Where(searchFor).Find(&accounts)
	if res.Error != nil {
		logger.Errorf("Failed finding all accounts: %v", res.Error)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Send result
	result := map[string]interface{}{}
	result["accounts"] = accounts
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(result); err != nil {
		logger.Errorf("Error encoding response: %v", err)
	}
}

func (self *RestEndpoint) validateAccount(account *model.Account) ([]JsonApiError, error) {
	errs := []JsonApiError{}
	if account.AgentID == "" {
		err := JsonApiError{
			Status: fmt.Sprintf("%d", http.StatusBadRequest),
			Title:  "Validation Error",
			Detail: "An agent id is required",
		}
		errs = append(errs, err)
	}
	if account.ServiceInstanceID == "" {
		err := JsonApiError{
			Status: fmt.Sprintf("%d", http.StatusBadRequest),
			Title:  "Validation Error",
			Detail: "A service instance id is required",
		}
		errs = append(errs, err)
	}
	if account.Email == "" {
		err := JsonApiError{
			Status: fmt.Sprintf("%d", http.StatusBadRequest),
			Title:  "Validation Error",
			Detail: "An email address is required",
		}
		errs = append(errs, err)
	} else {
		_, err := mail.ParseAddress(account.Email)
		if err != nil {
			errs = append(errs, JsonApiError{
				Status: fmt.Sprintf("%d", http.StatusBadRequest),
				Title:  "Validation Error",
				Detail: fmt.Sprintf("A valid email address is required (%v)", err),
			})
		}
	}
	if account.Name == "" {
		err := JsonApiError{
			Status: fmt.Sprintf("%d", http.StatusBadRequest),
			Title:  "Validation Error",
			Detail: "A name is required",
		}
		errs = append(errs, err)
	} else if !strings.HasPrefix(account.Email, account.Name+"@") {
		err := JsonApiError{
			Status: fmt.Sprintf("%d", http.StatusBadRequest),
			Title:  "Validation Error",
			Detail: "The name must match the first part of the email address",
		}
		errs = append(errs, err)
	}
	if account.DisplayName == "" {
		err := JsonApiError{
			Status: fmt.Sprintf("%d", http.StatusBadRequest),
			Title:  "Validation Error",
			Detail: "A display name is required",
		}
		errs = append(errs, err)
	}
	// Default the domain foreign key
	if account.DomainID == "" {
		domainName := strings.Split(account.Email, "@")[1]
		var domain model.Domain
		searchFor := &model.Domain{Name: domainName, AgentID: account.AgentID}
		if res := self.db.Where(searchFor).Limit(1).First(&domain); res.RecordNotFound() {
			err := JsonApiError{
				Status: fmt.Sprintf("%d", http.StatusBadRequest),
				Title:  "Validation Error",
				Detail: "A valid domain is required",
			}
			errs = append(errs, err)
		} else if res.Error != nil {
			logger.Errorf("Failed validating domain: %v", res.Error)
			err := JsonApiError{
				Status: fmt.Sprintf("%d", http.StatusInternalServerError),
				Title:  "Validation Error",
				Detail: "An internal error has occurred",
			}
			errs = append(errs, err)
		} else {
			account.DomainID = domain.ID
		}
	} else {
		var domain model.Domain
		searchFor := &model.Domain{ID: account.DomainID, AgentID: account.AgentID}
		if res := self.db.Where(searchFor).Limit(1).First(&domain); res.RecordNotFound() {
			err := JsonApiError{
				Status: fmt.Sprintf("%d", http.StatusBadRequest),
				Title:  "Validation Error",
				Detail: "A valid domain reference is required",
			}
			errs = append(errs, err)
		} else if res.Error != nil {
			logger.Errorf("Failed validating domain: %v", res.Error)
			err := JsonApiError{
				Status: fmt.Sprintf("%d", http.StatusInternalServerError),
				Title:  "Validation Error",
				Detail: "An internal error has occurred",
			}
			errs = append(errs, err)
		}
	}
	return errs, nil
}
