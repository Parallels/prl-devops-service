# Database Module

GORM-based database abstraction layer for prl-devops-service.

## Quick Start

### Installation

Dependencies are already added to go.mod:
```bash
cd src
go mod download
```

### SQLite Example

```go
import (
    "github.com/Parallels/prl-devops-service/database"
    "github.com/Parallels/prl-devops-service/database/common"
)

cfg := common.Config{
    Type: common.SQLite,
    SQLite: common.SQLiteConfig{
        StoragePath: "./data",
        FileName:    "app.db",
    },
    Debug: true,
}

db, err := database.Initialize(cfg)
if err != nil {
    log.Fatal(err)
}
defer db.Close()
```

### PostgreSQL Example

```go
cfg := common.Config{
    Type: common.PostgreSQL,
    PostgreSQL: common.PostgreSQLConfig{
        Host:     "localhost",
        Port:     5432,
        Database: "prl_devops",
        Username: "postgres",
        Password: "password",
    },
}

db, err := database.Initialize(cfg)
```

## Features

- ✅ SQLite and PostgreSQL support
- ✅ Connection pooling
- ✅ Health checks
- ✅ Error mapping
- ✅ Base models with timestamps
- ✅ Singleton pattern
- ✅ Transaction support

## Documentation

- [Full Integration Guide](../../docs/database-integration.md)
- [Example Code](../../docs/examples/database_example.go)
- [GORM Documentation](https://gorm.io/docs/)

## Structure

```
database/
├── common/           # Shared types
│   ├── config.go    # Database configuration
│   ├── base_model.go
│   ├── base_store.go
│   └── errors.go
├── interfaces/      # Interface definitions
│   ├── service.go
│   └── store.go
└── service/         # Implementation
    ├── service.go
    ├── sqlite.go
    └── postgres.go
```

## Next Steps

See [database-integration.md](../../docs/database-integration.md) for:
- Migration support
- Store implementations
- Advanced features
- Testing strategies

## Testing

```bash
# Build module
go build ./database/...

# Run example
cd docs/examples
go run database_example.go
```

## Issues

This addresses GitHub issue [#418](https://github.com/Parallels/prl-devops-service/issues/418).

See issues #419-427 for next steps in the database module implementation.
