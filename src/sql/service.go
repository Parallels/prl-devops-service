package sql

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"fmt"
	"os"
)

type MySQLAuthConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	Database string
}

type MySQLService struct{}

func (a *MySQLService) Connect() (*sql.DB, error) {
	config := MySQLAuthConfig{
		Host:     getEnv("MYSQL_HOST", "localhost"),
		Port:     getEnv("MYSQL_PORT", "3306"),
		User:     getEnv("MYSQL_USER", ""),
		Password: getEnv("MYSQL_PASSWORD", ""),
		Database: getEnv("MYSQL_DATABASE", ""),
	}

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s", config.User, config.Password, config.Host, config.Port, config.Database)
	return sql.Open("mysql", dsn)
}

func GenerateId() string {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		panic(err)
	}
	return hex.EncodeToString(bytes)
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}
