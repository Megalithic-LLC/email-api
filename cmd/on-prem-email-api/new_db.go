package main

import (
	"fmt"
	"net/url"
	"os"
	"strings"

	"github.com/on-prem-net/email-api/model"
	"github.com/docktermj/go-logger/logger"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	_ "github.com/jinzhu/gorm/dialects/postgres"
)

var (
	mysqlEnvVarNames = []string{
		"CLEARDB_DATABASE_URL",
		"JAWSDB_MARIA_URL",
		"JAWSDB_URL",
	}
	postgresEnvVarNames = []string{
		"DATABASE_URL",
		"CITUS_URL",
	}
)

type MyGormLoggerType struct{}

func (self MyGormLoggerType) Print(v ...interface{}) {
	if len(v) >= 4 && v[0] == "sql" {
		duration := fmt.Sprintf("%v", v[2])
		sql := fmt.Sprintf("%v", v[3])
		sql = strings.Replace(sql, "`", "", -1)
		sql = strings.Replace(sql, "  ", " ", -1)
		sql = strings.Replace(sql, " = ?", "=?", -1)
		logger.Debugf("[sql] %s %s", duration, sql)
	} else {
		line := gorm.LogFormatter(v...)
		logger.Debug(line)
	}
}

func newDB() *gorm.DB {
	var parsedURL *url.URL

	var dialect string

	// Support env based configuration of MySQL at Heroku, Deis, Dokku, etc
	for _, envVarName := range mysqlEnvVarNames {
		urlString := os.Getenv(envVarName)
		if urlString != "" {
			var err error
			parsedURL, err = url.Parse(urlString)
			if err != nil {
				logger.Fatalf("Failed parsing %s: %v", envVarName, err)
				return nil
			}
			dialect = "mysql"
			break
		}
	}

	// Support env based configuration of Postgres at Heroku, Deis, Dokku, etc
	for _, envVarName := range postgresEnvVarNames {
		urlString := os.Getenv(envVarName)
		if urlString != "" {
			var err error
			parsedURL, err = url.Parse(urlString)
			if err != nil {
				logger.Fatalf("Failed parsing %s: %v", envVarName, err)
				return nil
			}
			dialect = "postgres"
			break
		}
	}

	// Provide a default for developers
	if parsedURL == nil {
		parsedURL, _ = url.Parse("mysql://root:@/megalithic")
		dialect = "mysql"
		logger.Debugf("No database URL defined; defaulting to %s", parsedURL.String())
	}

	// Force some properties required for MySQL
	if dialect == "mysql" {
		query := parsedURL.Query()
		if query.Get("charset") == "" {
			query.Add("charset", "utf8")
		}
		query.Set("parseTime", "True")
		query.Del("reconnect")
		parsedURL.RawQuery = query.Encode()
	}

	var openString string
	switch dialect {
	case "mysql":
		openString = generateMysqlOpenString(parsedURL)
	case "postgres":
		openString = generatePostgresOpenString(parsedURL)
	}

	// Connect to DB
	db, err := gorm.Open(dialect, openString)
	if err != nil {
		logger.Fatalf("Failed connecting to %s database: %v", dialect, err)
		return nil
	}

	// Configure logging
	var mylogger MyGormLoggerType
	db.LogMode(true)
	db.SetLogger(mylogger)

	logger.Infof("Attached to %s at %s", dialect, parsedURL.String())

	// Perform migrations
	for _, mymodel := range model.AllModels {
		db.AutoMigrate(mymodel)
	}
	logger.Infof("Completed migrations")

	return db
}

func generateMysqlOpenString(dbUrl *url.URL) string {
	openString := ""
	if dbUrl.User != nil {
		openString = dbUrl.User.String()
	}
	openString += "@"
	if dbUrl.Host != "" {
		openString += fmt.Sprintf("(%s)", dbUrl.Host)
	}
	openString += dbUrl.Path
	if dbUrl.RawQuery != "" {
		openString += "?"
		openString += dbUrl.RawQuery
	}
	return openString
}

func generatePostgresOpenString(dbUrl *url.URL) string {
	params := map[string]string{
		"host": dbUrl.Hostname(),
	}
	if port := dbUrl.Port(); port != "" {
		params["port"] = port
	}
	if dbUrl.User != nil {
		params["user"] = dbUrl.User.Username()
		if password, hasPassword := dbUrl.User.Password(); hasPassword {
			params["password"] = password
		}
	}
	params["dbname"] = dbUrl.Path[1:]

	array := []string{}
	for k, v := range params {
		array = append(array, fmt.Sprintf("%s=%s", k, v))
	}
	return strings.Join(array, " ")
}
