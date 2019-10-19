package restendpoint

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/Megalithic-LLC/on-prem-email-api/model"
	"github.com/docktermj/go-logger/logger"
	"github.com/go-redis/redis"
	"github.com/gorilla/context"
	"github.com/jinzhu/gorm"
)

var (
	ignorePrefixes = []string{
		"/v1/confirmEmails",
		"/v1/token",
	}
)

type AuthenticationMiddleware struct {
	db          *gorm.DB
	redisClient *redis.Client
}

func NewAuthenticationMiddleware(db *gorm.DB, redisClient *redis.Client) *AuthenticationMiddleware {
	self := AuthenticationMiddleware{
		db:          db,
		redisClient: redisClient,
	}
	return &self
}

func (self *AuthenticationMiddleware) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		logger.Tracef("AuthenticationMiddleware(%s)", req.URL.Path)

		// Allow certain paths
		for _, ignorePrefix := range ignorePrefixes {
			if strings.HasPrefix(req.URL.Path, ignorePrefix) {
				next.ServeHTTP(w, req)
				return
			}
		}

		// Don't try to authenticate new user registrations
		if req.Method == "POST" && req.URL.Path == "/v1/users" {
			next.ServeHTTP(w, req)
			return
		}

		// Allow known bearer tokens
		authHeader := req.Header.Get("Authorization")
		if strings.HasPrefix(authHeader, "Bearer ") {

			bearerTokenString := authHeader[7:]

			// Support agent tokens
			if _, err := self.redisClient.Get(fmt.Sprintf("tok:%s", bearerTokenString)).Result(); err != nil && err != redis.Nil {
				logger.Errorf("Failed looking up agent token: %v", err)
				sendInternalServerError(w)
				return
			} else if err == nil {
				bearerToken, err := parseTokenString(bearerTokenString)
				if err != nil {
					logger.Errorf("Failed parsing bearer token: %v", err)
					sendInternalServerError(w)
					return
				}
				logger.Debugf("User %s authorized via agent key", bearerToken.UserID)
				context.Set(req, "currentAgentID", bearerToken.AgentID)
				context.Set(req, "currentUserID", bearerToken.UserID)
				next.ServeHTTP(w, req)
				return
			}

			// Support API Keys
			var apiKey model.ApiKey
			searchFor := &model.ApiKey{Key: bearerTokenString}
			if res := self.db.Where(searchFor).Limit(1).First(&apiKey); res.Error == nil {
				logger.Debugf("User %s authorized via API Key", apiKey.CreatedByUserID)
				context.Set(req, "currentUserID", apiKey.CreatedByUserID)
				next.ServeHTTP(w, req)
				return
			} else if res.Error != nil && !res.RecordNotFound() {
				logger.Errorf("Failed finding API Key: %v", res.Error)
				sendInternalServerError(w)
				return
			}

		}

		// Allow agents who are streaming just to wait to get claimed
		if values := req.Header["X-Agentid"]; len(values) > 0 {
			agentID := values[0]
			// TODO validate in DB
			context.Set(req, "currentAgentID", agentID)
			next.ServeHTTP(w, req)
			return
		}

		http.Error(w, "Forbidden", http.StatusForbidden)
	})
}
