package main

import (
	"github.com/docktermj/go-logger/logger"
	"github.com/Megalithic-LLC/on-prem-email-api/model"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	"golang.org/x/crypto/bcrypt"
)

const AdminUserID = "blmkmfd5jj89vu275l3g"

type AdminUserCreator struct{}

func newAdminUserCreator(db *gorm.DB) *AdminUserCreator {
	logger.Tracef("newAdminUserCreator()")

	var user model.User
	searchFor := &model.User{ID: AdminUserID}
	if res := db.Where(searchFor).First(&user); res.RecordNotFound() {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte("password"), bcrypt.DefaultCost)
		if err != nil {
			logger.Fatalf("Failed generating password for new admin user: %v", err)
		}

		user = model.User{
			ID:       AdminUserID,
			Username: "admin",
			Password: hashedPassword,
			First:    "Admin",
		}
		if err := db.Create(&user).Error; err != nil {
			logger.Fatalf("Failed creating new admin user: %v", err)
		}
	} else if res.Error != nil {
		logger.Fatalf("Failed looking up admin user: %v", res.Error)
		return nil
	}

	return nil
}
