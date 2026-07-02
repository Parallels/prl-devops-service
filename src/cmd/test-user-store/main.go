package main

import (
	"context"
	"fmt"

	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/database/stores"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func main() {
	fmt.Println("Testing user store database integration...")

	// Connect to database
	db, err := gorm.Open(sqlite.Open("data/database.db"), &gorm.Config{})
	if err != nil {
		fmt.Printf("Failed to connect: %v\n", err)
		return
	}

	// Initialize store
	stdCtx := context.Background()
	userStore := stores.GetUserDataStoreInstance()
	if err := userStore.Init(stdCtx, db); err != nil {
		fmt.Printf("Failed to init store: %v\n", err)
		return
	}

	// Test operations
	baseCtx := basecontext.NewRootBaseContext()

	// Get root user
	user, diag := userStore.GetUserByUsername(*baseCtx, "root")
	if diag.HasErrors() {
		fmt.Printf("❌ GetUserByUsername failed: %v\n", diag.Errors)
		return
	}

	fmt.Printf("✅ Found user:\n")
	fmt.Printf("   ID: %s\n", user.ID)
	fmt.Printf("   Username: %s\n", user.Username)
	fmt.Printf("   Name: %s\n", user.Name)
	fmt.Printf("   Email: %s\n", user.Email)

	// List all users
	users, diag := userStore.GetUsers(*baseCtx)
	if diag.HasErrors() {
		fmt.Printf("❌ GetUsers failed: %v\n", diag.Errors)
		return
	}

	fmt.Printf("\n✅ Total users: %d\n", len(users))

	fmt.Println("\n🎉 User store integration working!")
}
