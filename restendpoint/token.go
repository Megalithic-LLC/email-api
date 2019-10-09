package restendpoint

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/docktermj/go-logger/logger"
	"github.com/go-redis/redis"
	"github.com/on-prem-net/email-api/model"
	"golang.org/x/crypto/bcrypt"
)

var (
	jwtSecret       = []byte("möpsi")
	tokenTTL        = time.Duration(15) * time.Minute
	refreshTokenTTL = time.Duration(24) * time.Hour
)

type TokenClaims struct {
	UserID  string `json:"user"`
	AgentID string `json:"agent,omitempty"`
	jwt.StandardClaims
}

type RefreshTokenClaims struct {
	UserID string `json:"user"`
	jwt.StandardClaims
}

type Credentials struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func (self *RestEndpoint) createToken(w http.ResponseWriter, req *http.Request) {
	logger.Tracef("RestEndpoint:createToken")

	// Receive username/password submission
	var credentials Credentials
	if err := json.NewDecoder(req.Body).Decode(&credentials); err != nil {
		logger.Errorf("Failed decoding credentials: %v", err)
		sendBadRequestError(w, err)
		return
	}

	// Validate username
	filterBy := &model.User{Username: credentials.Username}
	var user model.User
	if res := self.db.Where(filterBy).First(&user); res.RecordNotFound() {
		logger.Warnf("No such user: %s", credentials.Username)
		w.WriteHeader(http.StatusUnauthorized)
		return
	} else if res.Error != nil {
		logger.Errorf("Failed looking up user: %v", res.Error)
		sendInternalServerError(w)
		return
	}

	// Validate password
	if err := bcrypt.CompareHashAndPassword(user.Password, []byte(credentials.Password)); err != nil {
		logger.Errorf("Wrong password attempt for user %s", user.Username)
		sendInternalServerError(w)
		return
	}

	self.generateAndStoreTokensThenRespond(user.ID, "", w, req)
}

func (self *RestEndpoint) refreshToken(w http.ResponseWriter, req *http.Request) {
	logger.Tracef("RestEndpoint:refreshToken()")

	// Receive username/password submission
	credentials := map[string]string{}
	if err := json.NewDecoder(req.Body).Decode(&credentials); err != nil {
		logger.Errorf("Failed decoding credentials: %v", err)
		sendBadRequestError(w, err)
		return
	}

	// Validate refresh token
	refreshTokenString := credentials["refresh_token"]
	if _, err := self.redisClient.Get(fmt.Sprintf("rtok:%s", refreshTokenString)).Result(); err != nil {
		if err == redis.Nil {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		logger.Errorf("Failed looking up refresh token: %v", err)
		sendInternalServerError(w)
		return
	}
	refreshToken, err := parseTokenString(refreshTokenString)
	if err != nil {
		logger.Errorf("Failed parsing refresh token: %v", err)
		sendInternalServerError(w)
		return
	}

	// Validate user is still active
	filterBy := &model.User{ID: refreshToken.UserID}
	var user model.User
	if res := self.db.Where(filterBy).First(&user); res.RecordNotFound() {
		logger.Warnf("No such user: %s", refreshToken.UserID)
		w.WriteHeader(http.StatusUnauthorized)
		return
	} else if res.Error != nil {
		logger.Errorf("Failed looking up user: %v", res.Error)
		sendInternalServerError(w)
		return
	}

	self.generateAndStoreTokensThenRespond(user.ID, "", w, req)
}

func (self *RestEndpoint) generateAndStoreTokensThenRespond(userID, agentID string, w http.ResponseWriter, req *http.Request) {
	logger.Tracef("RestEndpoint:generateAndStoreTokensThenRespond(%s)", userID)

	// Generate token
	tokenString, err := self.generateTokenString(tokenTTL, userID, agentID)
	if err != nil {
		logger.Errorf("Failed signing token: %v", err)
		sendInternalServerError(w)
		return
	}

	// Generate refresh token
	refreshTokenString, err := self.generateRefreshTokenString(refreshTokenTTL, fmt.Sprintf("%v", userID))
	if err != nil {
		logger.Errorf("Failed signing refresh token: %v", err)
		sendInternalServerError(w)
		return
	}

	// Store token and refresh token
	pipeline := self.redisClient.Pipeline()
	pipeline.Set(fmt.Sprintf("tok:%v", tokenString), "1", tokenTTL)
	pipeline.Set(fmt.Sprintf("rtok:%v", refreshTokenString), "1", refreshTokenTTL)
	if _, err := pipeline.Exec(); err != nil {
		logger.Errorf("Failed storing tokens: %v", err)
		sendInternalServerError(w)
		return
	}

	// Respond
	type jwtToken struct {
		Token        string `json:"token"`
		RefreshToken string `json:"refresh_token"`
	}
	json.NewEncoder(w).Encode(jwtToken{Token: tokenString, RefreshToken: refreshTokenString})
}

func (self *RestEndpoint) generateTokenString(ttl time.Duration, userID, agentID string) (string, error) {
	expirationTime := time.Now().Add(ttl)
	claims := &TokenClaims{
		UserID:  userID,
		AgentID: agentID,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expirationTime.Unix(),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecret)
}

func (self *RestEndpoint) generateRefreshTokenString(ttl time.Duration, userID string) (string, error) {
	expirationTime := time.Now().Add(ttl)
	claims := &RefreshTokenClaims{
		UserID: userID,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expirationTime.Unix(),
		},
	}
	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return refreshToken.SignedString(jwtSecret)
}
func parseTokenString(tokenString string) (*TokenClaims, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
		}
		return jwtSecret, nil
	})
	if err != nil {
		return nil, err
	}
	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		tokenClaims := TokenClaims{
			UserID: claims["user"].(string),
		}
		if agent, ok := claims["agent"]; ok {
			tokenClaims.AgentID = agent.(string)
		}
		return &tokenClaims, nil
	} else {
		return nil, errors.New("Invalid token claims")
	}

}
