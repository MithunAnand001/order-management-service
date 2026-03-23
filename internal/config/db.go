package config

import (
	"fmt"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// DBConnector defines the interface for establishing database connections.
type DBConnector interface {
	EstablishPostgresConnection(dbURL string) (*gorm.DB, error)
}

// Database encapsulates the GORM instance.
type Database struct {
	Instance *gorm.DB
}

// DBService is the concrete implementation of the DBConnector interface.
type DBService struct{}

// EstablishPostgresConnection opens a new connection to the PostgreSQL database.
func (s *DBService) EstablishPostgresConnection(dbURL string) (*gorm.DB, error) {
	db, err := gorm.Open(postgres.Open(dbURL), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Verify connection
	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}

	if err := sqlDB.Ping(); err != nil {
		return nil, err
	}

	return db, nil
}

// GetDSN forms the PostgreSQL connection string.
func GetDSN(cfg *Config) string {
	return fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		cfg.DBHost, cfg.DBPort, cfg.DBUser, cfg.DBPassword, cfg.DBName)
}
