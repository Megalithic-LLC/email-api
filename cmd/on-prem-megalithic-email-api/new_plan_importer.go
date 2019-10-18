package main

import (
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/Megalithic-LLC/on-prem-email-api/model"
	"github.com/docktermj/go-logger/logger"
	"github.com/jinzhu/gorm"
	"github.com/rakyll/statik/fs"
)

type PlanImporter struct {
	privatefs http.FileSystem
	db        *gorm.DB
}

func newPlanImporter(db *gorm.DB, privatefs http.FileSystem) *PlanImporter {
	logger.Tracef("newPlanImporter()")
	self := PlanImporter{db: db, privatefs: privatefs}
	if err := self.importPlans(); err != nil {
		logger.Fatalf("Failed importing plans: %v", err)
	}
	return &self
}

func (self PlanImporter) importPlans() error {
	return fs.Walk(self.privatefs, "/plans", func(path string, fileInfo os.FileInfo, err error) error {
		extension := filepath.Ext(fileInfo.Name())
		if strings.HasPrefix(path, "/plans") && extension == ".json" && strings.Count(path, "/") == 2 {
			contents, err := fs.ReadFile(self.privatefs, path)
			if err != nil {
				return err
			}
			var plan model.Plan
			if err := json.Unmarshal(contents, &plan); err != nil {
				return err
			}

			if err := self.db.Where(model.Plan{ID: plan.ID}).Assign(plan).FirstOrCreate(&plan).Error; err != nil {
				logger.Errorf("Failed storing plan: %v", err)
				return err
			}
		}
		return nil
	})
}
