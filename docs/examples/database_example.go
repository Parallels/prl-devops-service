package main

import (
	"context"
	"fmt"
	"log"

	"github.com/Parallels/prl-devops-service/database"
	"github.com/Parallels/prl-devops-service/database/common"
	"gorm.io/gorm"
)

// Example model using BaseModel
type User struct {
	common.BaseModel
	Username string `gorm:"uniqueIndex;not null;type:text"`
	Email    string `gorm:"uniqueIndex;not null;type:text"`
	FullName string `gorm:"type:text"`
}

func main() {
	// Example 1: SQLite Configuration
	sqliteConfig := common.Config{
		Type: common.SQLite,
		SQLite: common.SQLiteConfig{
			StoragePath: "./data",
			FileName:    "prl-devops.db",
		},
		Debug: true,
		Pool:  common.DefaultPoolConfig(),
	}

	// Example 2: PostgreSQL Configuration (uncomment to use)
	/*
		postgresConfig := common.Config{
			Type: common.PostgreSQL,
			PostgreSQL: common.PostgreSQLConfig{
				Host:     "localhost",
				Port:     5432,
				Database: "prl_devops",
				Username: "postgres",
				Password: "password",
				SSLMode:  false,
			},
			Debug: true,
			Pool:  common.DefaultPoolConfig(),
		}
	*/

	// Use SQLite for this example
	cfg := sqliteConfig
	// Uncomment to use PostgreSQL instead:
	// cfg := postgresConfig

	// Initialize database service
	dbService, err := database.Initialize(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer dbService.Close()

	// Get database connection
	db := dbService.GetDB()

	// Auto-migrate the User table
	if err := db.AutoMigrate(&User{}); err != nil {
		log.Fatalf("Failed to migrate: %v", err)
	}

	// Create a user
	user := User{
		BaseModel: common.BaseModel{
			ID:   "user-123",
			Slug: "john-doe",
		},
		Username: "johndoe",
		Email:    "john@example.com",
		FullName: "John Doe",
	}

	result := db.Create(&user)
	if result.Error != nil {
		// Check for specific errors
		if common.IsRecordNotFound(result.Error) {
			log.Println("Record not found")
		}
		log.Printf("Error creating user: %v", common.MapError(result.Error))
	} else {
		fmt.Printf("Created user: %s\n", user.Username)
	}

	// Find user by username
	var foundUser User
	if err := db.Where("username = ?", "johndoe").First(&foundUser).Error; err != nil {
		if common.IsRecordNotFound(err) {
			log.Println("User not found")
		} else {
			log.Printf("Error finding user: %v", err)
		}
	} else {
		fmt.Printf("Found user: %s (%s)\n", foundUser.Username, foundUser.Email)
	}

	// Update user
	foundUser.FullName = "John Smith"
	if err := db.Save(&foundUser).Error; err != nil {
		log.Printf("Error updating user: %v", err)
	} else {
		fmt.Printf("Updated user: %s\n", foundUser.FullName)
	}

	// List all users
	var users []User
	if err := db.Find(&users).Error; err != nil {
		log.Printf("Error listing users: %v", err)
	} else {
		fmt.Printf("Total users: %d\n", len(users))
		for _, u := range users {
			fmt.Printf("  - %s (%s)\n", u.Username, u.Email)
		}
	}

	// Health check
	ctx := context.Background()
	if err := dbService.Health(ctx); err != nil {
		log.Printf("Health check failed: %v", err)
	} else {
		fmt.Println("Database health check: OK")
	}

	// Transaction example
	err = db.Transaction(func(tx *gorm.DB) error {
		// Create multiple users in a transaction
		users := []User{
			{
				BaseModel: common.BaseModel{ID: "user-456", Slug: "jane-doe"},
				Username:  "janedoe",
				Email:     "jane@example.com",
				FullName:  "Jane Doe",
			},
			{
				BaseModel: common.BaseModel{ID: "user-789", Slug: "bob-smith"},
				Username:  "bobsmith",
				Email:     "bob@example.com",
				FullName:  "Bob Smith",
			},
		}

		for _, u := range users {
			if err := tx.Create(&u).Error; err != nil {
				return err
			}
		}

		return nil
	})

	if err != nil {
		log.Printf("Transaction failed: %v", err)
	} else {
		fmt.Println("Transaction completed successfully")
	}

	fmt.Println("\nDatabase integration example completed!")
}
