package restendpoint

import (
	"context"
	"net"
	"net/http"
	"os"

	"github.com/Megalithic-LLC/on-prem-email-api/agentstreamendpoint"
	"github.com/docktermj/go-logger/logger"
	"github.com/go-redis/redis"
	"github.com/gorilla/mux"
	"github.com/jinzhu/gorm"
)

type RestEndpoint struct {
	agentStreamEndpoint *agentstreamendpoint.AgentStreamEndpoint
	authMiddleware      AuthenticationMiddleware
	db                  *gorm.DB
	listener            net.Listener
	redisClient         *redis.Client
	router              *mux.Router
	server              *http.Server
}

func New(
	agentStreamEndpoint *agentstreamendpoint.AgentStreamEndpoint,
	authMiddleware *AuthenticationMiddleware,
	db *gorm.DB,
	redisClient *redis.Client,
) *RestEndpoint {

	router := mux.NewRouter()

	self := RestEndpoint{
		agentStreamEndpoint: agentStreamEndpoint,
		authMiddleware:      *authMiddleware,
		db:                  db,
		redisClient:         redisClient,
		router:              router,
	}

	router.HandleFunc("/v1/accounts", self.getAccounts).Methods("GET")
	router.HandleFunc("/v1/accounts", self.createAccount).Methods("POST")
	router.HandleFunc("/v1/accounts/{id}", self.deleteAccount).Methods("DELETE")
	router.HandleFunc("/v1/accounts/{id}", self.getAccount).Methods("GET")

	router.HandleFunc("/v1/agentStream", self.agentStream).Methods("GET")

	router.HandleFunc("/v1/agents", self.getAgents).Methods("GET")
	router.HandleFunc("/v1/agents", self.createAgent).Methods("POST")
	router.HandleFunc("/v1/agents/{id}", self.getAgent).Methods("GET")
	router.HandleFunc("/v1/agents/{id}", self.createAgent).Methods("PUT")

	router.HandleFunc("/v1/apiKeys", self.getApiKeys).Methods("GET")
	router.HandleFunc("/v1/apiKeys", self.createApiKey).Methods("POST")
	router.HandleFunc("/v1/apiKeys/{id}", self.deleteApiKey).Methods("DELETE")
	router.HandleFunc("/v1/apiKeys/{id}", self.getApiKey).Methods("GET")

	router.HandleFunc("/v1/confirmEmails/{id}", self.getConfirmEmail).Methods("GET")

	router.HandleFunc("/v1/domains", self.getDomains).Methods("GET")
	router.HandleFunc("/v1/domains", self.createDomain).Methods("POST")
	router.HandleFunc("/v1/domains/{id}", self.deleteDomain).Methods("DELETE")
	router.HandleFunc("/v1/domains/{id}", self.getDomain).Methods("GET")

	router.HandleFunc("/v1/endpoints", self.getEndpoints).Methods("GET")
	router.HandleFunc("/v1/endpoints", self.createEndpoint).Methods("POST")
	router.HandleFunc("/v1/endpoints/{id}", self.deleteEndpoint).Methods("DELETE")
	router.HandleFunc("/v1/endpoints/{id}", self.getEndpoint).Methods("GET")

	router.HandleFunc("/v1/plans", self.getPlans).Methods("GET")
	router.HandleFunc("/v1/plans/{id}", self.getPlan).Methods("GET")

	router.HandleFunc("/v1/snapshots", self.getSnapshots).Methods("GET")
	router.HandleFunc("/v1/snapshots", self.createSnapshot).Methods("POST")
	router.HandleFunc("/v1/snapshots/{id}", self.deleteSnapshot).Methods("DELETE")
	router.HandleFunc("/v1/snapshots/{id}", self.getSnapshot).Methods("GET")

	router.HandleFunc("/v1/tokenAuth", self.createToken).Methods("POST")
	router.HandleFunc("/v1/tokenRefresh", self.refreshToken).Methods("POST")

	router.HandleFunc("/v1/users", self.createUser).Methods("POST")
	router.HandleFunc("/v1/users/{id}", self.getUser).Methods("GET")

	router.Use(self.authMiddleware.Middleware)

	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}

	http.Handle("/", router)
	listener, err := net.Listen("tcp", "0.0.0.0:"+port)
	if err != nil {
		logger.Fatalf("Failed listening: %v", err)
	}
	self.server = &http.Server{}

	go func() {
		if err := self.server.Serve(listener); err != nil {
			logger.Errorf("Failed listening: %v", err)
		}
	}()

	logger.Infof("Listening for http on port %d", listener.Addr().(*net.TCPAddr).Port)
	return &self
}

func (self *RestEndpoint) Shutdown(ctx context.Context) {
	self.server.Shutdown(ctx)
	logger.Infof("Rest endpoint shutdown")
}
