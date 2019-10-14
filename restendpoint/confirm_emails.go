package restendpoint

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/docktermj/go-logger/logger"
	"github.com/go-redis/redis"
	"github.com/gorilla/mux"
	"github.com/Megalithic-LLC/on-prem-email-api/model"
)

func (self *RestEndpoint) getConfirmEmail(w http.ResponseWriter, req *http.Request) {
	id := mux.Vars(req)["id"]
	logger.Tracef("RestEndpoint:getConfirmEmail(%s)", id)

	// Find
	userAsJson, err := self.redisClient.Get(fmt.Sprintf("newuser:%v", id)).Result()
	if err != nil {
		if err == redis.Nil {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		logger.Errorf("Failed looking up unregistered user: %v", err)
		sendInternalServerError(w)
		return
	}

	// Create user
	var user model.User
	if err := json.Unmarshal([]byte(userAsJson), &user); err != nil {
		logger.Errorf("Failed unmarshaling unregistered user stored in Redis: %v", err)
		sendInternalServerError(w)
		return
	}
	if err := self.db.Create(&user).Error; err != nil {
		logger.Errorf("Failed creating new user after email confirmed: %v", err)
		sendInternalServerError(w)
		return
	}

	// Since user has been created, remove invite from Redis so that future clicks on the
	// link in the email don't try to create new users
	if _, err := self.redisClient.Del(fmt.Sprintf("newuser:%v", id)).Result(); err != nil {
		logger.Errorf("Failed deleting new user invite from redis: %v", err)
	}

	// Send result
	result := map[string]interface{}{}
	result["confirmEmail"] = map[string]interface{}{"id": id}
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(result); err != nil {
		logger.Errorf("Error encoding response: %v", err)
	}
}
