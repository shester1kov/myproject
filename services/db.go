package services

import (
	"log"
	"project/models"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

func InitDB() {
	dsn := "host=62.76.233.254 user=student password=67 dbname=new_test_store port=5432 sslmode=disable"
	var err error
	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	err = DB.AutoMigrate(&models.Category{}, &models.Product{}, &models.User{}, &models.Order{}, &models.OrderProduct{}, &models.Review{})
	if err != nil {
		log.Fatalf("Migration failed: %v", err)
	}
}
