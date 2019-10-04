package main

import (
	"net/http"

	"github.com/docktermj/go-logger/logger"
	_ "github.com/on-prem-net/email-api/statik"
	"github.com/rakyll/statik/fs"
)

func newPrivateFs() http.FileSystem {
	if privatefs, err := fs.New(); err != nil {
		logger.Fatalf("Failed creating private fs: %v", err)
		return nil
	} else {
		return privatefs
	}
}
