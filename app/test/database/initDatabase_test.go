package database_test

import (
	"testing"

	"github.com/sd0ric4/book-reader-backend/app/config"
	"github.com/sd0ric4/book-reader-backend/app/database"
)

func TestInitMySQL(t *testing.T) {
	// Mock configuration
	config.Config = &config.ConfigStruct{
		MySQL: config.MySQLConfig{
			User:     "root",
			Password: "root",
			Host:     "192.168.0.118",
			Port:     3306,
			DBName:   "bookdb",
		},
	}

	database.InitMySQL()
}
