package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/database/models"
	"github.com/Parallels/prl-devops-service/database/stores"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type JSONData struct {
	Schema struct {
		Version string `json:"version"`
	} `json:"schema"`
	Claims []models.Claim `json:"claims"`
	Roles  []models.Role  `json:"roles"`
	Users  []models.User  `json:"users"`
}

func main() {
	jsonFile := flag.String("json", "out/binaries/data.json", "Path to JSON data file")
	dbPath := flag.String("db", "data/database.db", "Path to SQLite database")
	force := flag.Bool("force", false, "Force recreation of database")
	flag.Parse()

	fmt.Println("🚀 Starting full data migration...")

	// Remove existing database if force flag is set
	if *force {
		fmt.Println("🗑️  Removing existing database...")
		os.Remove(*dbPath)
	}

	// Initialize database
	stdCtx := context.Background()
	db, err := gorm.Open(sqlite.Open(*dbPath), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		fmt.Printf("❌ Failed to connect to database: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("✅ Database connected")

	// Auto-migrate
	fmt.Println("📦 Running auto-migrations...")
	if err := db.AutoMigrate(
		&models.User{},
		&models.Role{},
		&models.Claim{},
		&models.ApiKey{},
		&models.Configuration{},
		&models.Activity{},
	); err != nil {
		fmt.Printf("❌ Auto-migration failed: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("✅ Migrations complete")

	// Read JSON file
	fmt.Println("📖 Reading JSON data...")
	data, err := os.ReadFile(*jsonFile)
	if err != nil {
		fmt.Printf("❌ Failed to read JSON file: %v\n", err)
		os.Exit(1)
	}

	var jsonData JSONData
	if err := json.Unmarshal(data, &jsonData); err != nil {
		fmt.Printf("❌ Failed to parse JSON: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("✅ JSON parsed: %d claims, %d roles, %d users\n",
		len(jsonData.Claims), len(jsonData.Roles), len(jsonData.Users))

	// Initialize stores
	claimStore := stores.GetClaimDataStoreInstance()
	if err := claimStore.Init(stdCtx, db); err != nil {
		fmt.Printf("❌ Failed to init claim store: %v\n", err)
		os.Exit(1)
	}

	roleStore := stores.GetRoleDataStoreInstance()
	if err := roleStore.Init(stdCtx, db); err != nil {
		fmt.Printf("❌ Failed to init role store: %v\n", err)
		os.Exit(1)
	}

	userStore := stores.GetUserDataStoreInstance()
	if err := userStore.Init(stdCtx, db); err != nil {
		fmt.Printf("❌ Failed to init user store: %v\n", err)
		os.Exit(1)
	}

	baseCtx := basecontext.NewRootBaseContext()

	// Migrate claims first
	fmt.Println("\n📊 Migrating claims...")
	for i, claim := range jsonData.Claims {
		_, diag := claimStore.CreateClaim(*baseCtx, &claim)
		if diag.HasErrors() {
			fmt.Printf("⚠️  Claim %s: %v\n", claim.ID, diag.GetErrors())
		} else {
			fmt.Printf("   [%d/%d] %s\n", i+1, len(jsonData.Claims), claim.Name)
		}
	}
	fmt.Println("✅ Claims migrated")

	// Migrate roles with claims
	fmt.Println("\n📊 Migrating roles...")
	for i, role := range jsonData.Roles {
		_, diag := roleStore.CreateRole(*baseCtx, &role)
		if diag.HasErrors() {
			fmt.Printf("⚠️  Role %s: %v\n", role.ID, diag.GetErrors())
		} else {
			fmt.Printf("   [%d/%d] %s (%d claims)\n", i+1, len(jsonData.Roles), role.Name, len(role.Claims))
		}
	}
	fmt.Println("✅ Roles migrated")

	// Migrate users with roles
	fmt.Println("\n📊 Migrating users...")
	for i, user := range jsonData.Users {
		// Store roles and claims to associate after user creation
		userRoles := user.Roles
		userClaims := user.Claims

		// Create user directly in DB to avoid password re-hashing
		// The password in JSON is already hashed
		userCopy := user
		userCopy.Roles = nil
		userCopy.Claims = nil

		if err := db.Create(&userCopy).Error; err != nil {
			fmt.Printf("⚠️  User %s: %v\n", user.Username, err)
			continue
		}

		// Associate roles
		if len(userRoles) > 0 {
			var dbRoles []models.Role
			for _, role := range userRoles {
				var dbRole models.Role
				if err := db.Where("id = ?", role.ID).First(&dbRole).Error; err == nil {
					dbRoles = append(dbRoles, dbRole)
				}
			}
			if len(dbRoles) > 0 {
				db.Model(&userCopy).Association("Roles").Append(dbRoles)
			}
		}

		// Associate claims
		if len(userClaims) > 0 {
			var dbClaims []models.Claim
			for _, claim := range userClaims {
				var dbClaim models.Claim
				if err := db.Where("id = ?", claim.ID).First(&dbClaim).Error; err == nil {
					dbClaims = append(dbClaims, dbClaim)
				}
			}
			if len(dbClaims) > 0 {
				db.Model(&userCopy).Association("Claims").Append(dbClaims)
			}
		}

		fmt.Printf("   [%d/%d] %s (%d roles, %d claims)\n", i+1, len(jsonData.Users),
			user.Username, len(userRoles), len(userClaims))
	}
	fmt.Println("✅ Users migrated")

	// Verify migration
	fmt.Println("\n✅ Verifying migration...")

	var userCount, roleCount, claimCount int64
	db.Model(&models.User{}).Count(&userCount)
	db.Model(&models.Role{}).Count(&roleCount)
	db.Model(&models.Claim{}).Count(&claimCount)

	fmt.Printf("   Users: %d\n", userCount)
	fmt.Printf("   Roles: %d\n", roleCount)
	fmt.Printf("   Claims: %d\n", claimCount)

	fmt.Println("\n🎉 Migration complete!")
	fmt.Println("Database is ready for use")
}
