package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/serviceprovider/dbservice"
)

func main() {
	fmt.Println("🧪 End-to-End Database Integration Test")
	fmt.Println("========================================\n")

	// Initialize database service
	ctx := basecontext.NewRootBaseContext()
	db, err := dbservice.InitDatabase(ctx)
	if err != nil {
		fmt.Printf("❌ Failed to initialize database: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("✅ Database initialized")

	// Test 1: Get all users
	fmt.Println("\n📋 Test 1: Get All Users")
	users, err := db.GetUsers(ctx, "")
	if err != nil {
		fmt.Printf("❌ Failed: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("✅ Success: Found %d users\n", len(users))
	for _, user := range users {
		fmt.Printf("   - %s (%s)\n", user.Username, user.Email)
	}

	// Test 2: Get specific user
	fmt.Println("\n📋 Test 2: Get User by Username")
	user, err := db.GetUser(ctx, "root")
	if err != nil {
		fmt.Printf("❌ Failed: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("✅ Success: Found user %s (ID: %s)\n", user.Username, user.ID)

	// Test 3: Get user roles
	fmt.Println("\n📋 Test 3: Get User Roles")
	roles, err := db.GetUserRoles(ctx, user.ID)
	if err != nil {
		fmt.Printf("❌ Failed: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("✅ Success: Found %d roles\n", len(roles))
	for _, role := range roles {
		fmt.Printf("   - %s\n", role.Name)
	}

	// Test 4: Get user claims
	fmt.Println("\n📋 Test 4: Get User Claims")
	claims, err := db.GetUserClaims(ctx, user.ID)
	if err != nil {
		fmt.Printf("❌ Failed: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("✅ Success: Found %d claims\n", len(claims))
	if len(claims) > 0 {
		fmt.Printf("   - First 5 claims:\n")
		for i, claim := range claims {
			if i >= 5 {
				break
			}
			fmt.Printf("     • %s (%s)\n", claim.Name, claim.Description)
		}
	}

	// Test 5: Get all roles
	fmt.Println("\n📋 Test 5: Get All Roles")
	allRoles, err := db.GetRoles(ctx, "")
	if err != nil {
		fmt.Printf("❌ Failed: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("✅ Success: Found %d roles\n", len(allRoles))
	for _, role := range allRoles {
		fmt.Printf("   - %s\n", role.Name)
	}

	// Test 6: Get all claims
	fmt.Println("\n📋 Test 6: Get All Claims")
	allClaims, err := db.GetClaims(ctx, "")
	if err != nil {
		fmt.Printf("❌ Failed: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("✅ Success: Found %d claims\n", len(allClaims))

	// Test 7: Get specific role
	fmt.Println("\n📋 Test 7: Get Specific Role")
	role, err := db.GetRole(ctx, "SUPER_USER")
	if err != nil {
		fmt.Printf("❌ Failed: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("✅ Success: Found role %s\n", role.Name)

	// Test 8: Get specific claim
	fmt.Println("\n📋 Test 8: Get Specific Claim")
	claim, err := db.GetClaim(ctx, "LIST_USER")
	if err != nil {
		fmt.Printf("❌ Failed: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("✅ Success: Found claim %s (%s)\n", claim.Name, claim.Description)

	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("🎉 All database tests passed!")
	fmt.Println(strings.Repeat("=", 60))

	// Check if API server is running for HTTP tests
	fmt.Println("\n📡 Checking if API server is running...")
	resp, err := http.Get("http://localhost:8080/api/v1/health")
	if err != nil {
		fmt.Println("⚠️  API server not running (this is OK for database-only tests)")
		fmt.Println("   To test API endpoints, start the server with: go run src/cmd/api/main.go")
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode == 200 {
		fmt.Println("✅ API server is running!")
		testAPIEndpoints()
	}
}

func testAPIEndpoints() {
	fmt.Println("\n🌐 Testing API Endpoints")
	fmt.Println("========================\n")

	// For now, just verify health endpoint
	fmt.Println("📋 Test API-1: Health Check")
	resp, err := http.Get("http://localhost:8080/api/v1/health")
	if err != nil {
		fmt.Printf("❌ Failed: %v\n", err)
		return
	}
	defer resp.Body.Close()

	var healthData map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&healthData)
	fmt.Printf("✅ Success: API is healthy\n")
	fmt.Printf("   Status: %d\n", resp.StatusCode)

	// Note: Full API endpoint tests would require authentication token
	fmt.Println("\n⚠️  Full API endpoint tests require authentication")
	fmt.Println("   Run 'curl' commands manually to test authenticated endpoints")
}
