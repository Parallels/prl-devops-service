package main
package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/data"
	"github.com/Parallels/prl-devops-service/database/stores"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func main() {
	jsonFile := flag.String("json", "out/binaries/data.json", "Path to JSON data file")
	dbPath := flag.String("db", "data/database.db", "Path to SQLite database")
	flag.Parse()

	fmt.Println("🚀 Starting database migration from JSON...")

	// Load JSON using existing data package
	rootCtx := basecontext.NewRootBaseContext()
	jsonDB := data.NewJsonDatabase(rootCtx, *jsonFile)
	if err := jsonDB.Load(rootCtx); err != nil {
		fmt.Printf("❌ Failed to load JSON: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("✅ JSON loaded")

	// Initialize database
	stdCtx := context.Background()
	db, err := gorm.Open(sqlite.Open(*dbPath), &gorm.Config{})
	if err != nil {
		fmt.Printf("❌ Failed to connect to database: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("✅ Database connected")

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

	// Get data from JSON
	claims, err := jsonDB.GetClaims(rootCtx, "")
	if err != nil {
		fmt.Printf("❌ Failed to get claims: %v\n", err)
		os.Exit(1)
	}

	roles, err := jsonDB.GetRoles(rootCtx, "")
	if err != nil {
		fmt.Printf("❌ Failed to get roles: %v\n", err)
		os.Exit(1)
	}

	users, err := jsonDB.GetUsers(rootCtx, "")
	if err != nil {
		fmt.Printf("❌ Failed to get users: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("📊 Found: %d claims, %d roles, %d users\n", len(claims), len(roles), len(users))

	// Migrate claims
	fmt.Println("\n📊 Migrating claims...")
	for i, claim := range claims {
		_, diag := claimStore.CreateClaim(*baseCtx, &claim)
		if diag.HasErrors() {
			fmt.Printf("⚠️  Claim %s: %v\n", claim.ID, diag.Errors)
		} else {
			fmt.Printf("   [%d/%d] %s\n", i+1, len(claims), claim.Name)
		}
	}
	fmt.Println("✅ Claims migrated")

	// Migrate roles
	fmt.Println("\n📊 Migrating roles...")
	for i, role := range roles {
		_, diag := roleStore.CreateRole(*baseCtx, &role)
		if diag.HasErrors() {
			fmt.Printf("⚠️  Role %s: %v\n", role.ID, diag.Errors)
		} else {
			fmt.Printf("   [%d/%d] %s\n", i+1, len(roles), role.Name)
		}
	}
	fmt.Println("✅ Roles migrated")

	// Migrate users
	fmt.Println("\n📊 Migrating users...")
	for i, user := range users {
		_, diag := userStore.CreateUser(*baseCtx, &user)
		if diag.HasErrors() {
			fmt.Printf("⚠️  User %s: %v\n", user.Username, diag.Errors)
		} else {
			fmt.Printf("   [%d/%d] %s (%s)\n", i+1, len(users), user.Username, user.Email)
		}
	}
	fmt.Println("✅ Users migrated")

	fmt.Println("\n🎉 Migration complete!")
	fmt.Println("Database is ready to use")
}
