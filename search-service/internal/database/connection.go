package database

import (
	"errors"
	"fmt"
	"log"
	"os"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func ConnectDB() (*gorm.DB, error) {
	dsn := os.Getenv("DSN")
	fmt.Println("dsn: ", dsn)
	if dsn == "" {
		return nil, errors.New("DSN can't left empty")
	}

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})

	if err != nil {
		return nil, err
	}

	log.Println("Connected to database")
	if err := startMigration(db); err != nil {
		return nil, err
	}

	return db, nil
}

func getRawSQLQuery(filePath string) (string, error) {
	f, err := os.ReadFile(filePath)
	return string(f), err
}

func ManualMigration(db *gorm.DB) error {
	filePath := "internal/database/migrations.sql"
	sqlQuery, err := getRawSQLQuery(filePath)
	if err != nil {
		return errors.New("Unable to read migrations.sql")
	}

	result := db.Exec(sqlQuery)
	if result.Error != nil {
		return result.Error
	}

	return nil
}

func startMigration(db *gorm.DB) error {
	// Auto migration
	if err := db.AutoMigrate(&Job{}); err != nil {
		return err
	}

	// Manual migration
	if err := ManualMigration(db); err != nil {
		return err
	}

	return nil
}
