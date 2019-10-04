package main

import (
	"net/url"
	"os"

	"github.com/docktermj/go-logger/logger"
	"github.com/go-redis/redis"
)

var (
	redisEnvVarNames = []string{
		"REDIS_URL",
		"OPENREDIS_URL",
		"REDISCLOUD_URL",
		"REDISGREEN_URL",
		"REDISTOGO_URL",
	}
)

func newRedisClient() *redis.Client {

	var parsedURL *url.URL

	// Support env based configuration of Redis at Heroku, Deis, Dokku, etc
	for _, envVarName := range redisEnvVarNames {
		urlString := os.Getenv(envVarName)
		if urlString != "" {
			var err error
			parsedURL, err = url.Parse(urlString)
			if err != nil {
				logger.Fatalf("Failed parsing %s: %v", envVarName, err)
				return nil
			}
			break
		}
	}

	// Provide a default for developers
	if parsedURL == nil {
		parsedURL, _ = url.Parse("redis://localhost:6379")
	}

	opts, err := redis.ParseURL(parsedURL.String())
	if err != nil {
		logger.Fatalf("Failed creating Redis connect options from URL: %v", err)
		return nil
	}

	client := redis.NewClient(opts)

	if _, err := client.Ping().Result(); err != nil {
		logger.Fatalf("Failed connecting to Redis: %v", err)
		return nil
	}

	logger.Infof("Attached to Redis at %s", opts.Addr)
	return client
}
