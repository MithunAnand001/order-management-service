package repository

import (
	"log"

	"order-management-service/internal/models"

	"gorm.io/gorm"
)

func RunMigrations(db *gorm.DB) {
	log.Println("Running database migrations...")
	// Enable uuid-ossp extension for gen_random_uuid()
	db.Exec("CREATE EXTENSION IF NOT EXISTS \"uuid-ossp\"")
	
	err := db.AutoMigrate(
		&models.User{},
		&models.Product{},
		&models.Order{},
		&models.OrderItem{},
		&models.OrderEventLog{},
	)
	if err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	// Seed Products using FirstOrCreate to avoid duplicates but ensure existence
	log.Println("Ensuring initial products exist...")
	products := []models.Product{
		{SKU: "PROD-001", Name: "Laptop", Description: "High-end gaming laptop", CurrentPrice: 1500.00, StockQuantity: 10},
		{SKU: "PROD-002", Name: "Mouse", Description: "Wireless ergonomic mouse", CurrentPrice: 50.00, StockQuantity: 100},
		{SKU: "PROD-003", Name: "Keyboard", Description: "Mechanical RGB keyboard", CurrentPrice: 100.00, StockQuantity: 50},
	}

	for _, p := range products {
		if err := db.Where(models.Product{SKU: p.SKU}).Attrs(p).FirstOrCreate(&p).Error; err != nil {
			log.Printf("Failed to ensure product %s: %v", p.SKU, err)
		}
	}

	log.Println("Migrations and seeding completed successfully")
}
