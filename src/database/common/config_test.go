package common
package common

import (
	"testing"
)

func TestParseDSN(t *testing.T) {
	tests := []struct {
		name         string
		dsn          string
		wantHost     string
		wantPort     int
		wantDatabase string
		wantUsername string
		wantPassword string
		wantSSLMode  bool
		wantErr      bool
	}{
		{
			name:         "Full DSN with password and SSL",
			dsn:          "postgresql://myuser:mypass@localhost:5432/mydb?sslmode=require",
			wantHost:     "localhost",
			wantPort:     5432,
			wantDatabase: "mydb",
			wantUsername: "myuser",
			wantPassword: "mypass",
			wantSSLMode:  true,
			wantErr:      false,
		},
		{
			name:         "DSN without password",
			dsn:          "postgresql://myuser@localhost:5432/mydb",
			wantHost:     "localhost",
			wantPort:     5432,
			wantDatabase: "mydb",
			wantUsername: "myuser",
			wantPassword: "",
			wantSSLMode:  false,
			wantErr:      false,
		},
		{
			name:         "DSN without port (default 5432)",
			dsn:          "postgresql://myuser:mypass@localhost/mydb",
			wantHost:     "localhost",
			wantPort:     5432,
			wantDatabase: "mydb",
			wantUsername: "myuser",
			wantPassword: "mypass",
			wantSSLMode:  false,
			wantErr:      false,
		},
		{
			name:         "DSN with verify-full SSL mode",
			dsn:          "postgresql://user:pass@db.example.com:5433/production?sslmode=verify-full",
			wantHost:     "db.example.com",
			wantPort:     5433,
			wantDatabase: "production",
			wantUsername: "user",
			wantPassword: "pass",
			wantSSLMode:  true,
			wantErr:      false,
		},
		{
			name:         "DSN with disable SSL mode",
			dsn:          "postgresql://admin:secret@192.168.1.100:5432/testdb?sslmode=disable",
			wantHost:     "192.168.1.100",
			wantPort:     5432,
			wantDatabase: "testdb",
			wantUsername: "admin",
			wantPassword: "secret",
			wantSSLMode:  false,
			wantErr:      false,
		},
		{
			name:         "Invalid DSN",
			dsn:          "not-a-valid-dsn",
			wantHost:     "",
			wantPort:     0,
			wantDatabase: "",
			wantUsername: "",
			wantPassword: "",
			wantSSLMode:  false,
			wantErr:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			host, port, database, username, password, sslMode, err := ParseDSN(tt.dsn)

			if (err != nil) != tt.wantErr {
				t.Errorf("ParseDSN() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr {
				return
			}

			if host != tt.wantHost {
				t.Errorf("ParseDSN() host = %v, want %v", host, tt.wantHost)
			}
			if port != tt.wantPort {
				t.Errorf("ParseDSN() port = %v, want %v", port, tt.wantPort)
			}
			if database != tt.wantDatabase {
				t.Errorf("ParseDSN() database = %v, want %v", database, tt.wantDatabase)
			}
			if username != tt.wantUsername {
				t.Errorf("ParseDSN() username = %v, want %v", username, tt.wantUsername)
			}
			if password != tt.wantPassword {
				t.Errorf("ParseDSN() password = %v, want %v", password, tt.wantPassword)
			}
			if sslMode != tt.wantSSLMode {
				t.Errorf("ParseDSN() sslMode = %v, want %v", sslMode, tt.wantSSLMode)
			}
		})
	}
}

func TestRedactDSNPassword(t *testing.T) {
	tests := []struct {
		name string
		dsn  string
		want string
	}{
		{
			name: "DSN with password",
			dsn:  "postgresql://myuser:mypassword@localhost:5432/mydb?sslmode=disable",
			want: "postgresql://myuser:***@localhost:5432/mydb?sslmode=disable",
		},
		{
			name: "DSN without password",
			dsn:  "postgresql://myuser@localhost:5432/mydb",
			want: "postgresql://myuser@localhost:5432/mydb",
		},
		{
			name: "DSN with special characters in password",
			dsn:  "postgresql://admin:p@ssw0rd!@db.example.com:5432/prod",
			want: "postgresql://admin:***@db.example.com:5432/prod",
		},
		{
			name: "Invalid DSN",
			dsn:  "not-a-valid-url",
			want: "postgresql://<invalid-dsn>",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := redactDSNPassword(tt.dsn)
			if got != tt.want {
				t.Errorf("redactDSNPassword() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPostgreSQLConfig_String(t *testing.T) {
	tests := []struct {
		name   string
		config PostgreSQLConfig
		want   string
	}{
		{
			name: "Config with DSN",
			config: PostgreSQLConfig{
				DSN: "postgresql://user:secret@localhost:5432/db",
			},
			want: "PostgreSQL(DSN=<redacted>)",
		},
		{
			name: "Config with individual fields",
			config: PostgreSQLConfig{
				Host:     "localhost",
				Port:     5432,
				Database: "testdb",
				Username: "admin",
				Password: "supersecret",
				SSLMode:  true,
			},
			want: "PostgreSQL(host=localhost, port=5432, database=testdb, username=admin, password=<redacted>, ssl=true)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.config.String()
			if got != tt.want {
				t.Errorf("PostgreSQLConfig.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPostgreSQLConfig_SafeDSN(t *testing.T) {
	tests := []struct {
		name   string
		config PostgreSQLConfig
		want   string
	}{
		{
			name: "Config with DSN containing password",
			config: PostgreSQLConfig{
				DSN: "postgresql://user:secret@localhost:5432/db?sslmode=require",
			},
			want: "postgresql://user:***@localhost:5432/db?sslmode=require",
		},
		{
			name: "Config with individual fields and SSL",
			config: PostgreSQLConfig{
				Host:     "db.example.com",
				Port:     5432,
				Database: "production",
				Username: "admin",
				Password: "supersecret",
				SSLMode:  true,
			},
			want: "postgresql://admin:***@db.example.com:5432/production?sslmode=require",
		},
		{
			name: "Config with individual fields without SSL",
			config: PostgreSQLConfig{
				Host:     "localhost",
				Port:     5432,
				Database: "testdb",
				Username: "testuser",
				Password: "testpass",
				SSLMode:  false,
			},
			want: "postgresql://testuser:***@localhost:5432/testdb?sslmode=disable",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.config.SafeDSN()
			if got != tt.want {
				t.Errorf("PostgreSQLConfig.SafeDSN() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPostgreSQLConfig_Validate(t *testing.T) {
	tests := []struct {
		name       string
		config     PostgreSQLConfig
		wantErrors bool
	}{
		{
			name: "Config with DSN - no validation needed",
			config: PostgreSQLConfig{
				DSN: "postgresql://user:pass@localhost:5432/db",
			},
			wantErrors: false,
		},
		{
			name: "Config with all fields set",
			config: PostgreSQLConfig{
				Host:     "localhost",
				Port:     5432,
				Database: "mydb",
				Username: "user",
				Password: "pass",
			},
			wantErrors: false,
		},
		{
			name: "Config with missing fields - should apply defaults",
			config: PostgreSQLConfig{
				// All empty - defaults should be applied
			},
			wantErrors: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			diag := tt.config.Validate()
			hasErrors := diag.HasErrors()
			if hasErrors != tt.wantErrors {
				t.Errorf("PostgreSQLConfig.Validate() errors = %v, want %v", hasErrors, tt.wantErrors)
			}
		})
	}
}
