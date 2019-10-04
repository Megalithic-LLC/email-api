package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/on-prem-net/email-api/agentstreamendpoint"
	"github.com/on-prem-net/email-api/restendpoint"
	"github.com/docktermj/go-logger/logger"
	"github.com/go-redis/redis"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	"github.com/karlkfi/inject"
)

var (
	graph inject.Graph

	adminUserCreator         *AdminUserCreator
	agentStreamEndpoint      *agentstreamendpoint.AgentStreamEndpoint
	authenticationMiddleware *restendpoint.AuthenticationMiddleware
	db                       *gorm.DB
	privateFs                http.FileSystem
	redisClient              *redis.Client
	restEndpoint             *restendpoint.RestEndpoint
	serviceImporter          *ServiceImporter
)

func init() {
	logger.SetLevel(logger.LevelTrace)
}

func main() {
	graph = inject.NewGraph()
	graph.Define(&adminUserCreator, inject.NewAutoProvider(newAdminUserCreator))
	graph.Define(&agentStreamEndpoint, inject.NewAutoProvider(agentstreamendpoint.New))
	graph.Define(&authenticationMiddleware, inject.NewAutoProvider(restendpoint.NewAuthenticationMiddleware))
	graph.Define(&db, inject.NewAutoProvider(newDB))
	graph.Define(&privateFs, inject.NewAutoProvider(newPrivateFs))
	graph.Define(&redisClient, inject.NewAutoProvider(newRedisClient))
	graph.Define(&restEndpoint, inject.NewAutoProvider(restendpoint.New))
	graph.Define(&serviceImporter, inject.NewAutoProvider(newServiceImporter))
	graph.ResolveAll()

	logger.Info("On-Prem Admin API started")

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)

	// Wait for shutdown
	<-c
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(5)*time.Second)
	defer cancel()
	restEndpoint.Shutdown(ctx)
	logger.Infof("Shutting down")
	if db != nil {
		db.Close()
	}
	os.Exit(0)
}
