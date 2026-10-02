package models

import (
	"fmt"
	"time"

	"taskhub/internal/configs"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

// InitDB initializes PostgreSQL connection with connection pooling and executes AutoMigrate
func InitDB(cfg *configs.Config) (*gorm.DB, error) {
	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=%s TimeZone=UTC",
		cfg.DBHost,
		cfg.DBUser,
		cfg.DBPassword,
		cfg.DBName,
		cfg.DBPort,
		cfg.DBSSLMode,
	)

	logLevel := gormlogger.Warn
	if cfg.AppEnv == "development" {
		logLevel = gormlogger.Info
	}

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: gormlogger.Default.LogMode(logLevel),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get sql.DB from gorm: %w", err)
	}

	// Set connection pooling parameters
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)

	// Execute AutoMigrate
	if err := AutoMigrate(db); err != nil {
		return nil, fmt.Errorf("failed to run database automigrations: %w", err)
	}

	return db, nil
}

// AutoMigrate runs schema auto-migrations for all core models
func AutoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&User{},
		&Project{},
		&Task{},
		&Comment{},
	)
}
