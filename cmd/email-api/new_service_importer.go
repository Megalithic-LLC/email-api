package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/docktermj/go-logger/logger"
	"github.com/on-prem-net/email-api/model"
	"github.com/jinzhu/gorm"
	"github.com/rakyll/statik/fs"
	"gopkg.in/russross/blackfriday.v2"
)

type ServiceImporter struct {
	privatefs http.FileSystem
	db        *gorm.DB
}

func newServiceImporter(db *gorm.DB, privatefs http.FileSystem) *ServiceImporter {
	logger.Tracef("newServiceImporter()")
	self := ServiceImporter{db: db, privatefs: privatefs}
	if err := self.importServices(); err != nil {
		logger.Fatalf("Failed importing services: %v", err)
	}
	return &self
}

func (self ServiceImporter) importServices() error {
	return fs.Walk(self.privatefs, "/services", func(path string, fileInfo os.FileInfo, err error) error {
		extension := filepath.Ext(fileInfo.Name())
		basename := path[0 : len(path)-len(extension)]
		if strings.HasPrefix(path, "/services") && extension == ".json" && strings.Count(path, "/") == 3 {
			contents, err := fs.ReadFile(self.privatefs, path)
			if err != nil {
				return err
			}
			var service model.Service
			if err := json.Unmarshal(contents, &service); err != nil {
				return err
			}

			longDescriptionPath := fmt.Sprintf("%s-longDescription.md", basename)
			if data, err := fs.ReadFile(self.privatefs, longDescriptionPath); err != nil {
				if !os.IsNotExist(err) {
					return err
				}
			} else {
				rendererParams := blackfriday.HTMLRendererParameters{Flags: blackfriday.HrefTargetBlank}
				renderer := blackfriday.NewHTMLRenderer(rendererParams)
				html := blackfriday.Run(data, blackfriday.WithNoExtensions(), blackfriday.WithRenderer(renderer))
				service.LongDescription = string(html)
			}

			if err := self.db.Where(model.Service{ID: service.ID}).Assign(service).FirstOrCreate(&service).Error; err != nil {
				logger.Errorf("Failed storing service: %v", err)
				return err
			}

			serviceDir := filepath.Dir(path)
			if err := self.importServicePlans(serviceDir, service); err != nil {
				logger.Errorf("Failed importing plans: %v", err)
				return err
			}
		}
		return nil
	})
}

func (self ServiceImporter) importServicePlans(dir string, service model.Service) error {
	logger.Tracef("ServiceImporter:importServicePlans(%s)", dir)
	return fs.Walk(self.privatefs, dir+"/plans", func(path string, fileInfo os.FileInfo, err error) error {
		extension := filepath.Ext(fileInfo.Name())
		if extension == ".json" {
			contents, err := fs.ReadFile(self.privatefs, path)
			if err != nil {
				return err
			}
			var plan model.Plan
			if err := json.Unmarshal(contents, &plan); err != nil {
				return err
			}
			plan.ServiceID = service.ID

			if err := self.db.Where(model.Plan{ID: plan.ID}).Assign(plan).FirstOrCreate(&plan).Error; err != nil {
				logger.Errorf("Failed storing plan: %v", err)
				return err
			}
		}
		return nil
	})
}
