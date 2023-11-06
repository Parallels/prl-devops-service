package sql

import (
	"database/sql"

	_ "github.com/go-sql-driver/mysql"
)

type DatabaseService interface {
	Connect() (*sql.DB, error)
}
