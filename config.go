package main

import (
	"fmt"
	"log"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// User model (must be defined here or imported
var DB *gorm.DB

func InitDB() {
	dsn := "host=localhost user=postgres password=root dbname=fileUpdate port=5432 sslmode=disable"
	var err error
	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("❌ Failed to connect to the database: %v", err)
	}

	// Run migrations
	err = DB.AutoMigrate(&User{})
	if err != nil {
		log.Fatalf("❌ Database migration failed: %v", err)
	}

	fmt.Println("✅ Database connected successfully!")
}
