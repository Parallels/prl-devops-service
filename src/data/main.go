package data

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/config"
	"github.com/Parallels/prl-devops-service/data/models"
	"github.com/Parallels/prl-devops-service/errors"
	"github.com/Parallels/prl-devops-service/helpers"
	"github.com/Parallels/prl-devops-service/security"

	"github.com/cjlapao/common-go/helper"
)

var (
	ErrDatabaseNotConnected = errors.NewWithCode("the database is not connected", 500)
	ErrNotAuthorized        = errors.NewWithCode("not authorized to view record", 403)
)

var (
	memoryDatabase    *JsonDatabase
	wg                = &sync.WaitGroup{}
	totalSaveRequests = 0
	mutexLock         sync.Mutex
)

type Data struct {
	Schema            models.DatabaseSchema     `json:"schema"`
	Configuration     *models.Configuration     `json:"configuration"`
	Users             []models.User             `json:"users"`
	Claims            []models.Claim            `json:"claims"`
	Roles             []models.Role             `json:"roles"`
	ApiKeys           []models.ApiKey           `json:"api_keys"`
	PackerTemplates   []models.PackerTemplate   `json:"virtual_machine_templates"`
	ManifestsCatalog  []models.CatalogManifest  `json:"catalog_manifests"`
	OrchestratorHosts []models.OrchestratorHost `json:"orchestrator_hosts"`
}

type JsonDatabase struct {
	ctx         basecontext.ApiContext
	Config      JsonDatabaseConfig
	connected   bool
	isSaving    bool
	saveProcess chan bool
	filename    string
	saveMutex   sync.Mutex
	cancel      chan bool
	data        Data
}

type JsonDatabaseConfig struct {
	DatabaseFilename    string        `json:"database_filename"`
	NumberOfBackupFiles int           `json:"number_of_backup_files"`
	SaveInterval        time.Duration `json:"save_interval"`
	BackupInterval      time.Duration `json:"backup_interval"`
}

func NewJsonDatabase(ctx basecontext.ApiContext, filename string) *JsonDatabase {
	if memoryDatabase != nil {
		return memoryDatabase
	}
	cfg := config.Get()

	memoryDatabase = &JsonDatabase{
		Config: JsonDatabaseConfig{
			DatabaseFilename:    filename,
			NumberOfBackupFiles: cfg.DbNumberBackupFiles(),
			SaveInterval:        cfg.DbSaveInterval(),
			BackupInterval:      cfg.DbBackupInterval(),
		},
		ctx:         ctx,
		connected:   false,
		isSaving:    false,
		filename:    filename,
		saveProcess: make(chan bool),
		data:        Data{},
	}

	wg = &sync.WaitGroup{}
	rootContext := basecontext.NewRootBaseContext()
	_ = memoryDatabase.Load(rootContext)

	memoryDatabase.cancel = make(chan bool)
	if err := memoryDatabase.SaveAsync(ctx); err != nil {
		ctx.LogErrorf("[Database] Error saving database: %v", err)
	}

	// Starting the automatic backup
	memoryDatabase.RunBackup(ctx)
	return memoryDatabase
}

func (j *JsonDatabase) Connect(ctx basecontext.ApiContext) error {
	ctx.LogDebugf("[Database] Connecting to database %s", j.filename)
	j.connected = true
	return nil
}

func (j *JsonDatabase) Load(ctx basecontext.ApiContext) error {
	ctx.LogInfof("[Database] Loading database from %s", j.filename)
	var data Data

	if _, err := os.Stat(j.filename); os.IsNotExist(err) {
		ctx.LogInfof("[Database] Database file does not exist, creating new file")
		file, err := os.Create(j.filename)
		if err != nil {
			ctx.LogErrorf("[Database] Error creating database file: %v", err)
			return err
		}
		if err := file.Close(); err != nil {
			return err
		}
	}

	file, err := os.Open(j.filename)
	if err != nil {
		ctx.LogErrorf("[Database] Error opening database file: %v", err)
		return err
	}

	defer file.Close()

	if err != nil {
		ctx.LogErrorf("[Database] Error getting database file info: %v", err)
		return err
	}

	isEmpty := false
	fileContent, err := helper.ReadFromFile(j.filename)
	if err != nil {
		isEmpty = true
	}
	if fileContent == nil || len(fileContent) == 0 {
		isEmpty = true
	}

	if isEmpty {
		ctx.LogInfof("[Database] Database file is empty, creating new file")
		j.data = Data{
			Users:            make([]models.User, 0),
			Claims:           make([]models.Claim, 0),
			Roles:            make([]models.Role, 0),
			ApiKeys:          make([]models.ApiKey, 0),
			PackerTemplates:  make([]models.PackerTemplate, 0),
			ManifestsCatalog: make([]models.CatalogManifest, 0),
		}

		err = j.SaveNow(ctx)
		if err != nil {
			ctx.LogErrorf("[Database] Error saving database file: %v", err)
			return err
		}

		j.connected = true
		return nil
	} else {
		ctx.LogInfof("[Database] Database file is not empty, loading data")

		// Backup the file before attempting to read it
		if err := j.Backup(ctx); err != nil {
			ctx.LogErrorf("[Database] Error managing backup files: %v", err)
		}

		content, err := helper.ReadFromFile(j.filename)
		if err != nil {
			ctx.LogErrorf("[Database] Error reading database file: %v", err)
			return err
		}
		if content == nil {
			ctx.LogErrorf("[Database] Error reading database file: %v", err)
			return err
		}

		// Trying to read the file unencrypted
		err = json.Unmarshal(content, &data)
		if err != nil {
			// Trying to read the file encrypted
			cfg := config.Get()
			if cfg.EncryptionPrivateKey() == "" {
				ctx.LogErrorf("[Database] Error reading database file: %v", err)
				return err
			}

			content, err := security.DecryptString(cfg.EncryptionPrivateKey(), content)
			if err != nil {
				ctx.LogErrorf("[Database] Error decrypting database file: %v", err)
				return err
			}

			err = json.Unmarshal([]byte(content), &data)
			if err != nil {
				ctx.LogErrorf("[Database] Error reading database file: %v", err)
				return err
			}
		}

		j.data = data
		j.connected = true
		return nil
	}
}

func (j *JsonDatabase) Disconnect(ctx basecontext.ApiContext) error {
	ctx.LogDebugf("[Database] Disconnecting from database")

	return nil
}

func (j *JsonDatabase) Filename() string {
	return j.filename
}

func (j *JsonDatabase) IsConnected() bool {
	return j.connected
}

func (j *JsonDatabase) SaveAs(ctx basecontext.ApiContext, filename string) error {
	oldFilename := j.filename
	baseDbDir := filepath.Dir(oldFilename)
	fileName := filepath.Base(filename)
	newFilename := filepath.Join(baseDbDir, fileName)

	ctx.LogDebugf("[Database] Saving database to %s", filename)
	j.filename = newFilename
	if !helper.FileExists(j.filename) {
		if _, err := os.Create(j.filename); err != nil {
			j.filename = oldFilename
			return err
		}
	}
	if err := j.processSave(ctx); err != nil {
		j.filename = oldFilename
		return err
	}

	j.filename = oldFilename
	return nil
}

func (j *JsonDatabase) SaveAsync(ctx basecontext.ApiContext) error {
	defer func() {
		if r := recover(); r != nil {
			ctx.LogErrorf("[Database] Panic occurred during save: %v", r)
		}
	}()

	go func() {
		// recover from panic
		defer func() {
			if r := recover(); r != nil {
				ctx.LogErrorf("[Database] Panic occurred during save: %v", r)
			}
		}()

		for {
			select {
			case <-j.cancel:
				return
			default:
				ctx.LogDebugf("[Database] Enqueuing next save request")
				time.Sleep(j.Config.SaveInterval)
				wg.Add(1)
				go j.ProcessSaveQueue(ctx)
				wg.Wait()
				ctx.LogDebugf("[Database] Save request completed")
			}
		}
	}()

	return nil
}

func (j *JsonDatabase) SaveNow(ctx basecontext.ApiContext) error {
	ctx.LogDebugf("[Database] Received for save request")
	if err := j.processSave(ctx); err != nil {
		ctx.LogErrorf("[Database] Error saving database: %v", err)
		return err
	}
	return nil
}

func (j *JsonDatabase) ProcessSaveQueue(ctx basecontext.ApiContext) {
	defer wg.Done()
	ctx.LogDebugf("[Database] Received for save request")
	mutexLock.Lock()
	if err := j.processSave(ctx); err != nil {
		ctx.LogErrorf("[Database] Error saving database: %v", err)
	}
	mutexLock.Unlock()
}

func (j *JsonDatabase) processSave(ctx basecontext.ApiContext) error {
	j.saveMutex.Lock()
	cfg := config.Get()
	ctx.LogDebugf("[Database] Saving database to %s", j.filename)
	j.isSaving = true
	if j.filename == "" {
		j.saveMutex.Unlock()
		return errors.NewWithCode("the database filename is not set", 500)
	}

	// Trying to open the file and waiting for it to be ready
	var file *os.File
	openCount := 0
	for {
		openCount++
		ctx.LogDebugf("[Database] Trying to open file %s, attempt %v", j.filename, openCount)
		var fileOpenError error
		file, fileOpenError = os.OpenFile(j.filename, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0o600)
		if fileOpenError == nil {
			ctx.LogDebugf("[Database] File %s opened successfully", j.filename)
			break
		}
	}

	defer file.Close()

	ctx.LogDebugf("[Database] File %s opened successfully", j.filename)
	jsonString, err := json.MarshalIndent(j.data, "", "  ")
	if err != nil {
		ctx.LogDebugf("[Database] Error marshalling data: %v", err)
		j.isSaving = false
		j.saveMutex.Unlock()
		return errors.NewFromError(err)
	}

	ctx.LogDebugf("[Database] Data marshalled successfully")
	if cfg.EncryptionPrivateKey() != "" {
		encJsonString, err := security.EncryptString(cfg.EncryptionPrivateKey(), string(jsonString))
		if err != nil {
			ctx.LogDebugf("[Database] Error encrypting data: %v", err)
			_, saveErr := file.Write(jsonString)
			if saveErr != nil {
				ctx.LogDebugf("[Database] Error writing data: %v", saveErr)
				j.isSaving = false
				j.saveMutex.Unlock()
				return errors.NewFromError(saveErr)
			}

			j.isSaving = false
			j.saveMutex.Unlock()
			return errors.NewFromError(err)
		}

		jsonString = encJsonString
	}

	ctx.LogDebugf("[Database] Writing data to file")
	_, err = file.Write(jsonString)
	if err != nil {
		ctx.LogDebugf("[Database] Error writing data: %v", err)
		j.isSaving = false
		j.saveMutex.Unlock()
		return err
	}

	if err := file.Close(); err != nil {
		ctx.LogDebugf("[Database] Error closing file: %v", err)
		j.isSaving = false
		j.saveMutex.Unlock()
		return err
	}

	ctx.LogDebugf("[Database] File %s saved successfully", j.filename)
	j.isSaving = false
	j.saveMutex.Unlock()
	return nil
}

func LockRecord(ctx basecontext.ApiContext, dbRecord *models.DbRecord) {
	mutexLock.Lock()
	if dbRecord == nil {
		dbRecord = &models.DbRecord{}
	}
	dbRecord.IsLocked = true
	dbRecord.LockedAt = helpers.GetUtcCurrentDateTime()
	dbRecord.LockedBy = ctx.GetUser().Email
	mutexLock.Unlock()
}

func UnlockRecord(ctx basecontext.ApiContext, dbRecord *models.DbRecord) {
	mutexLock.Lock()
	dbRecord.IsLocked = false
	mutexLock.Unlock()
}
