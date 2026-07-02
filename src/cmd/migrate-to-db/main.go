package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/data/models"
	"github.com/Parallels/prl-devops-service/database/stores"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type JSONData struct {
	Schema struct {
		Version string `json:"version"`
	} `json:"schema"`
	Configuration     *models.Configuration     `json:"configuration"`
	Users             []models.User             `json:"users"`
	Claims            []models.Claim            `json:"claims"`
	Roles             []models.Role             `json:"roles"`
	ApiKeys           []models.ApiKey           `json:"api_keys"`
	PackerTemplates   []models.PackerTemplate   `json:"virtual_machine_templates"`
	ManifestsCatalog  []models.CatalogManifest  `json:"catalog_manifests"`
	OrchestratorHosts []models.OrchestratorHost `json:"orchestrator_hosts"`
}

func main() {
	jsonFile := flag.String("json", "out/binaries/data.json", "Path to JSON data file")
	dbPath := flag.String("db", "data/database.db", "Path to SQLite database")
	flag.Parse()

	fmt.Println("🚀 Starting database migration...")
	fmt.Printf("   JSON: %s\n", *jsonFile)
	fmt.Printf("   DB:   %s\n", *dbPath)

	// Create data directory if it doesn't exist
	dbDir := filepath.Dir(*dbPath)
	if err := os.MkdirAll(dbDir, 0755); err != nil {
		fmt.Printf("❌ Failed to create directory: %v\n", err)
		os.Exit(1)
	}

	// Initialize database
	db, err := gorm.Open(sqlite.Open(*dbPath), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		fmt.Printf("❌ Failed to connect to database: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("✅ Database connected")

	// Auto-migrate all models
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
	fmt.Printf("✅ JSON parsed: %d users, %d roles, %d claims\n",
		len(jsonData.Users), len(jsonData.Roles), len(jsonData.Claims))

	// Initialize stores with context.Context
	stdCtx := context.Background()

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

	// Use basecontext for operations
	baseCtx := basecontext.NewRootBaseContext()

	// Migrate claims first
	fmt.Println("\n📊 Migrating claims...")
	for i, claim := range jsonData.Claims {
		_, diag := claimStore.CreateClaim(*baseCtx, &claim)
		if diag.HasErrors() {
			fmt.Printf("⚠️  Claim %s: %v\n", claim.ID, diag.Errors)
		} else {
			fmt.Printf("   [%d/%d] %s\n", i+1, len(jsonData.Claims), claim.Name)
		}
	}
	fmt.Println("✅ Claims migrated")

	// Migrate roles
	fmt.Println("\n📊 Migrating roles...")
	for i, role := range jsonData.Roles {
		_, diag := roleStore.CreateRole(*baseCtx, &role)
		if diag.HasErrors() {
			fmt.Printf("⚠️  Role %s: %v\n", role.ID, diag.Errors)
		} else {
			fmt.Printf("   [%d/%d] %s\n", i+1, len(jsonData.Roles), role.Name)
		}
	}
	fmt.Println("✅ Roles migrated")

	// Migrate users
	fmt.Println("\n📊 Migrating users...")
	for i, user := range jsonData.Users {
		_, diag := userStore.CreateUser(*baseCtx, &user)
		if diag.HasErrors() {
			fmt.Printf("⚠️  User %s: %v\n", user.Username, diag.Errors)
		} else {
			fmt.Printf("   [%d/%d] %s (%s)\n", i+1, len(jsonData.Users), user.Username, user.Email)
		}
	}
	fmt.Println("✅ Users migrated")

	// Backup JSON file
	backupPath := *jsonFile + ".backup"
	if err := os.Rename(*jsonFile, backupPath); err != nil {
		fmt.Printf("⚠️  Failed to backup JSON: %v\n", err)
	} else {
		fmt.Printf("✅ JSON backed up to %s\n", backupPath)
	}

	fmt.Println("\n🎉 Migration complete!")
	fmt.Println("   Run your application and it will use the database instead of JSON")
}
