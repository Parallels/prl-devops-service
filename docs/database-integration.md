# Database Module Integration

## Overview

This document describes the GORM-based database module that has been integrated into the prl-devops-service project. The module provides a clean abstraction for database operations supporting both SQLite and PostgreSQL backends.

## Structure

```
src/database/
├── database.go           # Main facade/entry point
├── common/               # Shared types and utilities
│   ├── config.go        # Database configuration
│   ├── base_model.go    # Base GORM models
│   ├── base_store.go    # Base data store functionality
│   └── errors.go        # Database error handling
├── interfaces/          # Interface definitions
│   ├── service.go       # Database service interface
│   └── store.go         # Store interface
└── service/             # Service implementation
    ├── service.go       # Main database service
    ├── sqlite.go        # SQLite initialization
    └── postgres.go      # PostgreSQL initialization
```

## Features

### ✅ Implemented

- [x] **GORM Integration**: Full GORM v1.31.2 support
- [x] **Dual Backend Support**: SQLite and PostgreSQL
- [x] **Connection Pooling**: Configurable connection pool settings
- [x] **Singleton Pattern**: Thread-safe singleton database service
- [x] **Health Checks**: Database connection health monitoring
- [x] **Error Mapping**: Database-specific error translation
- [x] **Base Models**: Reusable base models with timestamps
- [x] **Store Pattern**: Interface-based store pattern for data access

### 📋 Dependencies Added

```
gorm.io/gorm v1.31.2
gorm.io/driver/sqlite v1.6.0
gorm.io/driver/postgres v1.6.0
modernc.org/sqlite v1.53.0
github.com/jackc/pgx/v5 v5.10.0
```

## Usage

### Basic Initialization

```go
package main

import (
    "context"
    "log"
    
    "github.com/Parallels/prl-devops-service/database"
    "github.com/Parallels/prl-devops-service/database/common"
)

func main() {
    // Configure database (SQLite example)
    cfg := common.Config{
        Type: common.SQLite,
        SQLite: common.SQLiteConfig{
            StoragePath: "./data",
            FileName:    "prl-devops.db",
        },
        Debug: true,
        Pool: common.PoolConfig{
            MaxIdleConns:    10,
            MaxOpenConns:    100,
            ConnMaxLifetime: 3600 * 1000000000, // 1 hour
        },
    }
    
    // Initialize database service
    dbService, err := database.Initialize(cfg)
    if err != nil {
        log.Fatalf("Failed to initialize database: %v", err)
    }
    defer dbService.Close()
    
    // Check health
    ctx := context.Background()
    if err := dbService.Health(ctx); err != nil {
        log.Fatalf("Database health check failed: %v", err)
    }
    
    // Get database connection for operations
    db := dbService.GetDB()
    
    // Use db for GORM operations...
}
```

### PostgreSQL Configuration

```go
cfg := common.Config{
    Type: common.PostgreSQL,
    PostgreSQL: common.PostgreSQLConfig{
        Host:     "localhost",
        Port:     5432,
        Database: "prl_devops",
        Username: "postgres",
        Password: "your_password",
        SSLMode:  false,
    },
    Debug: true,
}
```

### Using the Singleton Pattern

```go
// Get existing instance
dbService := database.GetInstance()

// Reset instance (testing only)
database.Reset()
```

## Configuration

### Configuration Structure

```go
type Config struct {
    Type           DatabaseType      // "sqlite" or "postgresql"
    SQLite         SQLiteConfig      // SQLite settings
    PostgreSQL     PostgreSQLConfig  // PostgreSQL settings
    Debug          bool              // Enable debug logging
    Migrate        bool              // Auto-run migrations
    MigrationsPath string            // Path to migration files
    Pool           PoolConfig        // Connection pool settings
}
```

### Connection Pool Settings

```go
type PoolConfig struct {
    MaxIdleConns    int           // Maximum idle connections
    MaxOpenConns    int           // Maximum open connections
    ConnMaxLifetime time.Duration // Maximum connection lifetime
}
```

## Base Models

The module provides base models for common fields:

```go
// Base model without tenant support
type BaseModel struct {
    ID        string    `gorm:"primarykey;type:text;not null"`
    Slug      string    `gorm:"not null;type:text"`
    CreatedBy string    `gorm:"type:text"`
    CreatedAt time.Time `gorm:"type:timestamp;not null;default:CURRENT_TIMESTAMP"`
    UpdatedBy string    `gorm:"type:text"`
    UpdatedAt time.Time `gorm:"type:timestamp;not null;default:CURRENT_TIMESTAMP"`
}

// Base model with tenant support
type BaseModelWithTenant struct {
    ID        string    `gorm:"primarykey;type:text;not null"`
    TenantID  string    `gorm:"not null;type:text;index"`
    Slug      string    `gorm:"not null;type:text"`
    CreatedBy string    `gorm:"type:text"`
    UpdatedBy string    `gorm:"type:text"`
    CreatedAt time.Time `gorm:"type:timestamp;not null;default:CURRENT_TIMESTAMP"`
    UpdatedAt time.Time `gorm:"type:timestamp;not null;default:CURRENT_TIMESTAMP"`
}
```

## Error Handling

The module provides error mapping for common database errors:

```go
import "github.com/Parallels/prl-devops-service/database/common"

// Check for specific errors
if common.IsRecordNotFound(err) {
    // Handle not found
}

// Map database errors to domain errors
err = common.MapError(dbErr)
```

Available error types:
- `ErrRecordNotFound`: Record not found
- `ErrDuplicateKey`: Unique constraint violation
- `ErrForeignKeyViolation`: Foreign key constraint violation
- `ErrDatabaseConnection`: Database connection error

## Next Steps

### 🔄 To Be Implemented

1. **Migration Support** (Task #419-421)
   - Copy migrations package
   - Implement migration runner
   - Create initial schema migrations

2. **Store Implementation** (Task #422-425)
   - Migrate existing data stores to GORM
   - Implement user store
   - Implement catalog store
   - Implement orchestrator store

3. **Advanced Features** (Task #426-427)
   - Implement filters package
   - Add observability plugins
   - Transaction support
   - Query optimization

## Testing

To verify the integration:

```bash
# Build database module
cd /Users/saikumar.peddireddy/Office/Repo/prl-devops-service/src
go build ./database/...

# Run tests (when implemented)
go test ./database/...
```

## Integration with Existing Code

The database module is designed to coexist with the existing JSON-based data layer. You can gradually migrate stores from `src/data` to use GORM while maintaining backward compatibility.

### Migration Pattern

1. Create new GORM-based store in `database/stores/`
2. Implement the `Store` interface
3. Register store with database service
4. Update business logic to use new store
5. Remove old JSON-based implementation

## References

- [GORM Documentation](https://gorm.io/docs/)
- [Original Issue #418](https://github.com/Parallels/prl-devops-service/issues/418)
- [Task Description](../../../go-backend-scaffolding/internal/database/task.md)

## Architecture Decisions

### Why GORM?

- Industry-standard Go ORM
- Support for multiple databases
- Rich feature set (migrations, associations, hooks)
- Active development and community support
- Type-safe query building

### Why Singleton Pattern?

- Ensures single database connection pool
- Simplifies dependency injection
- Thread-safe initialization
- Easy testing with Reset() function

### Why Store Pattern?

- Clear separation of concerns
- Easier testing with interfaces
- Allows gradual migration
- Supports dependency injection

## Troubleshooting

### SQLite Compilation Issues

If you encounter SQLite compilation errors, ensure you have a C compiler installed:
- macOS: `xcode-select --install`
- Linux: `apt-get install build-essential`
- Windows: Install MinGW-w64

The module uses `modernc.org/sqlite` which is a pure-Go SQLite driver, eliminating the need for CGO in most cases.

### Connection Pool Issues

If you experience connection pool exhaustion:
1. Increase `MaxOpenConns` in pool configuration
2. Ensure connections are properly closed
3. Check for connection leaks in your code
4. Monitor connection usage with health checks

## License

This module follows the same license as the prl-devops-service project.
