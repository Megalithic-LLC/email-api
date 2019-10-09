package restendpoint

import (
	"fmt"
	"net/http"
	"strings"

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
			if _, err := self.redisClient.Get(fmt.Sprintf("tok:%s", bearerTokenString)).Result(); err != nil {
				logger.Errorf("Failed looking up token: %v", err)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			bearerToken, err := parseTokenString(bearerTokenString)
			if err != nil {
				logger.Errorf("Failed parsing bearer token: %v", err)
				sendInternalServerError(w)
				return
			}
			context.Set(req, "currentUserID", bearerToken.UserID)
			next.ServeHTTP(w, req)
			return
		}

		// Allow agents who are streaming just to wait for claiming
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
