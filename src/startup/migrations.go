package startup

import (
	"fmt"
	"log"

	"github.com/Parallels/pd-api-service/serviceprovider"

	_ "github.com/go-sql-driver/mysql"
)

func ExecuteMigrations() {
	fmt.Println("Executing migrations")
	dbService := serviceprovider.Get().MySqlService
	if dbService == nil {
		log.Fatal("Error connecting to database")
	}

	db, err := dbService.Connect()
	if err != nil {
		log.Fatal(err)
	}

	// Execute the create table script
	_, err = db.Exec(`
  CREATE TABLE IF NOT EXISTS users (
    id VARCHAR(255) NOT NULL,
    name VARCHAR(255) NOT NULL,
    email VARCHAR(255) NOT NULL,
    password TEXT NOT NULL,
    PRIMARY KEY (id),
    UNIQUE (email)
);
    `)
	if err != nil {
		log.Fatal(err)
	}

	defer db.Close()

	fmt.Println("Table created successfully")
}
