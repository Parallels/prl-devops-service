package services

import (
	"Parallels/pd-api-service/data"
	"Parallels/pd-api-service/helpers"
	"Parallels/pd-api-service/models"
	sql_database "Parallels/pd-api-service/sql"
	"os"
	"strings"

	"github.com/cjlapao/common-go/commands"
	"github.com/cjlapao/common-go/log"
)

type Services struct {
	RunningUser      string
	Logger           *log.Logger
	ParallelsService *ParallelsService
	GitService       *GitService
	PackerService    *PackerService
	MySqlService     *sql_database.MySQLService
	JsonDatabase     *data.JsonDatabase
}

var globalServices *Services

func InitServices() {
	// Connect to the SQL server
	// dbService, err := initDatabase()
	// if err != nil {
	// 	panic(err)
	// }

	// Create a new Services struct and add the DB service
	globalServices = &Services{
		Logger: log.Get().WithTimestamp(),
	}
	stdout, err := helpers.ExecuteWithNoOutput(helpers.Command{
		Command: "whoami",
	})
	if err != nil {
		panic(err)
	}
	globalServices.RunningUser = strings.ReplaceAll(strings.TrimSpace(stdout), "\n", "")
	globalServices.ParallelsService = NewParallelsService()
	globalServices.GitService = NewGitService()
	globalServices.PackerService = NewPackerService()

	if globalServices.RunningUser == "root" {
		dbLocation := "/etc/parallels-api-service"
		err := helpers.CreateDirIfNotExist(dbLocation)
		if err != nil {
			panic(err)
		}
		globalServices.JsonDatabase = data.NewJsonDatabase(dbLocation + "/data.json")
		globalServices.JsonDatabase.Connect()
		globalServices.Logger.Info("Running as %s, using %s/data.json file", globalServices.RunningUser, dbLocation)
	} else {
		dbLocation := "/Users/" + globalServices.RunningUser + "/.parallels-api-service"
		err := helpers.CreateDirIfNotExist(dbLocation)
		if err != nil {
			panic(err)
		}
		globalServices.JsonDatabase = data.NewJsonDatabase(dbLocation + "/data.json")
		globalServices.JsonDatabase.Connect()
		globalServices.Logger.Info("Running as %s, using %s/data.json file", globalServices.RunningUser, dbLocation)
	}
}

func GetServices() *Services {
	return globalServices
}

func initDatabase() (*sql_database.MySQLService, error) {
	service := sql_database.MySQLService{}
	_, err := service.Connect()
	if err != nil {
		return nil, err
	}

	return &service, nil
}

func GetSystemUsers() ([]models.SystemUser, error) {
	result := make([]models.SystemUser, 0)
	out, err := commands.ExecuteWithNoOutput("dscl", ".", "list", "/Users")
	if err != nil {
		return nil, err
	}

	users := strings.Split(out, "\n")
	for _, user := range users {
		user = strings.TrimSpace(user)
		if user == "" {
			continue
		}
		userHomeDir := "/Users/" + user
		if _, err := os.Stat(userHomeDir); os.IsNotExist(err) {
			continue
		} else {
			result = append(result, models.SystemUser{
				Username: user,
				Home:     userHomeDir,
			})
		}
	}

	return result, nil
}
