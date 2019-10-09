package restendpoint

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/smtp"
	"os"
	"time"

	"github.com/docktermj/go-logger/logger"
	"github.com/gorilla/context"
	"github.com/gorilla/mux"
	"github.com/on-prem-net/email-api/model"
	"github.com/rs/xid"
	"golang.org/x/crypto/bcrypt"
)

func (self *RestEndpoint) createUser(w http.ResponseWriter, req *http.Request) {
	logger.Tracef("RestEndpoint:createUser()")

	// Decode request
	type createUserRequestType struct {
		Fields map[string]string `json:"user"`
	}
	var createUserRequest createUserRequestType
	if err := json.NewDecoder(req.Body).Decode(&createUserRequest); err != nil {
		logger.Errorf("Failed decoding user: %v", err)
		sendBadRequestError(w, err)
		return
	}
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(createUserRequest.Fields["password"]), bcrypt.DefaultCost)
	if err != nil {
		logger.Errorf("Failed encrypting password: %v", err)
		sendInternalServerError(w)
		return
	}
	user := &model.User{
		First:    createUserRequest.Fields["first"],
		Last:     createUserRequest.Fields["last"],
		Email:    createUserRequest.Fields["email"],
		Username: createUserRequest.Fields["username"],
		Password: hashedPassword,
	}

	// Validate
	if validationErrors, err := self.validateUser(user); err != nil {
		logger.Errorf("Failure validating user: %v", err)
		sendInternalServerError(w)
		return
	} else if len(validationErrors) > 0 {
		sendErrors(w, validationErrors)
		return
	}

	// Store temporarily in Redis
	if user.ID == "" {
		user.ID = xid.New().String()
	}
	userAsJson, err := json.Marshal(user)
	if err != nil {
		logger.Errorf("Failure marshalling user: %v", err)
		sendInternalServerError(w)
		return
	}
	inviteID := xid.New().String()
	ttl := time.Duration(4) * time.Hour
	if _, err := self.redisClient.Set(fmt.Sprintf("newuser:%v", inviteID), string(userAsJson), ttl).Result(); err != nil {
		logger.Errorf("Failed storing user: %v", err)
		sendInternalServerError(w)
		return
	}

	// Send a confirmation email
	smtpUser, smtpPass := os.Getenv("SENDGRID_USERNAME"), os.Getenv("SENDGRID_PASSWORD")
	if smtpUser == "" || smtpPass == "" {
		logger.Errorf("No SMTP provider configured; unable to send welcome email; user is unable to register")
		sendInternalServerError(w)
		return
	}
	{
		consoleURL := os.Getenv("CONSOLE_URL")
		if consoleURL == "" {
			consoleURL = "http://localhost:4200"
		}

		addr := "smtp.sendgrid.net:587"
		auth := smtp.PlainAuth("", smtpUser, smtpPass, "smtp.sendgrid.net")
		from := "postmaster@on-prem.net"
		to := []string{user.Email}
		msg := "" +
			"From: On-Prem.net <postmaster@on-prem.net>\r\n" +
			"To: \"" + user.First + " " + user.Last + "\" <" + user.Email + ">\r\n" +
			"Subject: Confirm your email address for on-prem.net\r\n" +
			"Content-Type: text/html\r\n" +
			"\r\n" +
			"<html><body>Please click <a href=\"" + consoleURL + "/confirm-email/" + inviteID + "\">here</a> " +
			"to confirm your email address.</body></html>"
		if err := smtp.SendMail(addr, auth, from, to, []byte(msg)); err != nil {
			logger.Errorf("Failed sending welcome email; user is unable to register: %v", err)
			sendInternalServerError(w)
			return
		}
		logger.Infof("Welcome email sent to %s", user.Email)
	}

	// Send result
	result := map[string]interface{}{}
	result["user"] = user
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(result); err != nil {
		logger.Errorf("Error encoding response: %v", err)
	}
}

func (self *RestEndpoint) getUser(w http.ResponseWriter, req *http.Request) {
	id := mux.Vars(req)["id"]
	logger.Tracef("RestEndpoint:getUser(%s)", id)

	currentUserID := context.Get(req, "currentUserID").(string)
	if id != currentUserID {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	// Find
	var user model.User
	{
		searchFor := &model.User{ID: id}
		if res := self.db.Where(searchFor).Limit(1).First(&user); res.RecordNotFound() {
			w.WriteHeader(http.StatusNotFound)
			return
		} else if res.Error != nil {
			logger.Errorf("Failed finding user: %v", res.Error)
			sendInternalServerError(w)
			return
		}
	}

	// Send result
	result := map[string]interface{}{}
	result["user"] = user
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(result); err != nil {
		logger.Errorf("Error encoding response: %v", err)
	}
}

func (self *RestEndpoint) validateUser(user *model.User) ([]JsonApiError, error) {
	errs := []JsonApiError{}
	if user.Email == "" {
		err := JsonApiError{
			Status: fmt.Sprintf("%d", http.StatusBadRequest),
			Title:  "Validation Error",
			Detail: "An email address is required",
		}
		errs = append(errs, err)
	}
	if user.Username == "" {
		err := JsonApiError{
			Status: fmt.Sprintf("%d", http.StatusBadRequest),
			Title:  "Validation Error",
			Detail: "A username is required",
		}
		errs = append(errs, err)
	}
	return errs, nil
}
