package data

import (
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/config"
	"github.com/Parallels/prl-devops-service/constants"
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
	ReverseProxy      *models.ReverseProxy      `json:"reverse_proxy"`
	ReverseProxyHosts []models.ReverseProxyHost `json:"reverse_proxy_hosts"`
}

type JsonDatabase struct {
	ctx         basecontext.ApiContext
	Config      JsonDatabaseConfig
	connected   bool
	isSaving    bool
	saveProcess chan bool
	filename    string
	saveMutex   sync.Mutex
	dataMutex   sync.RWMutex
	cancel      chan bool
	data        Data
}

type JsonDatabaseConfig struct {
	DatabaseFilename    string        `json:"database_filename"`
	NumberOfBackupFiles int           `json:"number_of_backup_files"`
	SaveInterval        time.Duration `json:"save_interval"`
	BackupInterval      time.Duration `json:"backup_interval"`
	AutoRecover         bool          `json:"auto_recover"`
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
			AutoRecover:         cfg.IsDatabaseAutoRecover(),
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
	if j.Config.AutoRecover {
		// recover from residual save files if any
		if recovered, err := j.recoverFromResidualSaveFiles(ctx, "*.save"); err != nil {
			ctx.LogErrorf("[Database] Error recovering from residual save files: %v", err)
			j.removeGlobFiles(ctx, "*.save")
		} else if recovered {
			return nil
		}
		// Recover from residual save files if any
		if recovered, err := j.recoverFromResidualSaveFiles(ctx, "*.save_bak"); err != nil {
			ctx.LogErrorf("[Database] Error recovering from residual save files: %v", err)
			j.removeGlobFiles(ctx, "*.save")
		} else if recovered {
			return nil
		}
		// Recover from crash files if any
		if recovered, err := j.recoverFromResidualSaveFiles(ctx, "*.panic"); err != nil {
			ctx.LogErrorf("[Database] Error recovering from panic file files: %v", err)
			j.removeGlobFiles(ctx, "*.panic")
		} else if recovered {
			return nil
		}
	}

	isEmpty, err := j.IsDataFileEmpty(ctx)
	if err != nil {
		ctx.LogErrorf("[Database] Error checking if database file is empty: %v", err)
		return err
	}

	if isEmpty {
		if err := j.loadFromEmpty(ctx); err != nil {
			ctx.LogErrorf("[Database] Error loading database file: %v", err)
			return err
		}
		return nil
	} else {
		if err := j.loadFromFile(ctx); err != nil {
			ctx.LogErrorf("[Database] Error loading database file: %v",
				err)
			return err
		}

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
	if cfg.GetRunningCommand() != constants.API_COMMAND && cfg.GetRunningCommand() != "" {
		ctx.LogDebugf("[Database] Skipping save request, command running: %s", cfg.GetRunningCommand())
		j.saveMutex.Unlock()
		return nil
	}

	ctx.LogDebugf("[Database] Saving database to %s", j.filename)
	j.isSaving = true
	if j.filename == "" {
		j.saveMutex.Unlock()
		return errors.NewWithCode("the database filename is not set", 500)
	}

	// Trying to open the file and waiting for it to be ready
	dateTimeForFile := time.Now().Format("20060102150405")
	tempFileName := j.filename + "." + dateTimeForFile + ".save"
	var tempFile *os.File
	openCount := 0
	maxOpenAttempts := 10
	for {
		openCount++
		ctx.LogDebugf("[Database] Trying to open file %s, attempt %v", tempFileName, openCount)
		var fileOpenError error
		tempFile, fileOpenError = os.OpenFile(tempFileName, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0o600)
		if fileOpenError == nil {
			ctx.LogDebugf("[Database] File %s opened successfully", j.filename)
			break
		}
		ctx.LogDebugf("[Database] Error opening file %s: %v", tempFileName, fileOpenError)
		if openCount > maxOpenAttempts {
			ctx.LogDebugf("[Database] Max attempts reached, aborting save")
			j.isSaving = false
			j.saveMutex.Unlock()
			return errors.NewFromError(fileOpenError)
		}

		time.Sleep(1 * time.Second)
	}

	defer tempFile.Close()

	j.dataMutex.RLock()
	jsonString, err := json.MarshalIndent(j.data, "", "  ")
	j.dataMutex.RUnlock()
	if err != nil {
		ctx.LogDebugf("[Database] Error marshalling data to temp file: %v", err)
		j.isSaving = false
		j.saveMutex.Unlock()
		return errors.NewFromError(err)
	}

	ctx.LogDebugf("[Database] Data marshalled successfully")
	// encrypting the data before saving it
	if cfg.EncryptionPrivateKey() != "" {
		encJsonString, err := security.EncryptString(cfg.EncryptionPrivateKey(), string(jsonString))
		if err != nil {
			ctx.LogDebugf("[Database] Error encrypting data: %v", err)
			_, saveErr := tempFile.Write(jsonString)
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
	_, err = tempFile.Write(jsonString)
	if err != nil {
		ctx.LogDebugf("[Database] Error writing data to temp file: %v", err)
		j.isSaving = false
		j.saveMutex.Unlock()
		return err
	}

	if err := tempFile.Close(); err != nil {
		ctx.LogDebugf("[Database] Error closing temp file: %v", err)
		j.isSaving = false
		j.saveMutex.Unlock()
		return err
	}

	// Copy the current file to a backup file
	if err = j.copyCurrentDbFileToTemp(ctx, dateTimeForFile); err != nil {
		ctx.LogDebugf("[Database] Error copying current file to backup: %v", err)
		j.isSaving = false
		return err
	}

	// Rename the temp file to the original filename
	err = os.Rename(tempFileName, j.filename)
	if err != nil {
		ctx.LogDebugf("[Database] Error renaming temp file: %v", err)
		j.isSaving = false
		j.saveMutex.Unlock()
		return err
	}

	// Delete the save backup temp file
	backupFilename := j.filename + "." + dateTimeForFile + ".save_bak"
	if helper.FileExists(backupFilename) {
		ctx.LogDebugf("[Database] Backup file %s exists, deleting it", backupFilename)

		err = os.Remove(backupFilename)
		if err != nil {
			ctx.LogDebugf("[Database] Error deleting temp file: %v", err)
			j.isSaving = false
			j.saveMutex.Unlock()
			return err
		}
	}

	ctx.LogDebugf("[Database] File %s saved successfully", j.filename)
	j.isSaving = false
	j.saveMutex.Unlock()
	return nil
}

func (j *JsonDatabase) copyCurrentDbFileToTemp(ctx basecontext.ApiContext, dateTimeForFile string) error {
	// Copy current file to a backup file
	backupFilename := j.filename + "." + dateTimeForFile + ".save_bak"
	inputFile, err := os.Open(j.filename)
	if err != nil {
		ctx.LogDebugf("[Database] Error opening file for backup: %v", err)
		j.isSaving = false
		j.saveMutex.Unlock()
		return err
	}
	defer inputFile.Close()

	outputFile, err := os.Create(backupFilename)
	if err != nil {
		ctx.LogDebugf("[Database] Error creating backup file: %v", err)
		j.isSaving = false
		j.saveMutex.Unlock()
		return err
	}
	defer outputFile.Close()

	_, err = io.Copy(outputFile, inputFile)
	if err != nil {
		ctx.LogDebugf("[Database] Error copying to backup file: %v", err)
		j.isSaving = false
		j.saveMutex.Unlock()
		return err
	}

	// Delete the original file
	err = os.Remove(j.filename)
	if err != nil {
		ctx.LogDebugf("[Database] Error deleting original file: %v", err)
		j.isSaving = false
		j.saveMutex.Unlock()
		return err
	}

	return nil
}
func IsRecordLocked(dbRecord *models.DbRecord) bool {
	if dbRecord == nil {
		return false
	}
	return dbRecord.IsLocked
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
	if dbRecord == nil {
		dbRecord = &models.DbRecord{}
	}
	dbRecord.IsLocked = false
	mutexLock.Unlock()
}

func (j *JsonDatabase) removeAllBackupSavedFiles(ctx basecontext.ApiContext, glob string) error {
	// Delete all *.save and *.save_bak files
	saveFiles, err := filepath.Glob(j.filename + glob)
	if err != nil {
		ctx.LogErrorf("[Database] Error finding save files: %v", err)
		return err
	}
	for _, file := range saveFiles {
		if err := os.Remove(file); err != nil {
			ctx.LogErrorf("[Database] Error deleting save file %s: %v", file, err)
			return err
		}
	}

	return nil
}

func (j *JsonDatabase) recoverFromResidualSaveFiles(ctx basecontext.ApiContext, glob string) (bool, error) {
	var data Data
	// Checking if there is any previous backing up saving files to recover from
	saveBackFiles, err := filepath.Glob(j.filename + glob)
	if err != nil {
		ctx.LogErrorf("[Database] Error finding save files: %v", err)
		return false, err
	}
	if len(saveBackFiles) > 0 {
		ctx.LogInfof("[Database] Found %d save files, attempting to recover the latest one", len(saveBackFiles))
		latestSaveFile := saveBackFiles[len(saveBackFiles)-1]
		content, err := helper.ReadFromFile(latestSaveFile)
		if err != nil {
			ctx.LogErrorf("[Database] Error reading save file %s: %v", latestSaveFile, err)
			return false, err
		}
		if len(content) == 0 {
			ctx.LogInfof("[Database] Save file %s is empty, ignoring", latestSaveFile)
			return false, nil
		}
		if err != nil {
			ctx.LogErrorf("[Database] Error unmarshalling save file %s: %v", latestSaveFile, err)
			return false, err
		}
		j.dataMutex.Lock()
		j.data = data
		j.connected = true
		j.dataMutex.Unlock()

		ctx.LogInfof("[Database] Successfully recovered data from save file %s", latestSaveFile)
		if err := j.SaveNow(ctx); err != nil {
			ctx.LogErrorf("[Database] Error saving database: %v", err)
			return false, err
		}

		if err := j.removeAllBackupSavedFiles(ctx, glob); err != nil {
			ctx.LogErrorf("[Database] Error removing backup files: %v", err)
			return false, err
		}

		return true, nil
	}

	return false, nil
}

func (j *JsonDatabase) recoverFromBackupFile(ctx basecontext.ApiContext) error {
	var data Data

	// Check if there are any backup files available
	backupFiles, err := filepath.Glob(j.filename + ".save.bak.*")
	if err != nil {
		ctx.LogErrorf("[Database] Error finding backup files: %v", err)
		return err
	}
	if len(backupFiles) > 0 {
		ctx.LogInfof("[Database] Found %d backup files, attempting to recover the latest one", len(backupFiles))
		latestBackupFile := backupFiles[len(backupFiles)-1]
		content, err := helper.ReadFromFile(latestBackupFile)
		if err != nil {
			ctx.LogErrorf("[Database] Error reading backup file %s: %v", latestBackupFile, err)
			return err
		}

		if len(content) == 0 {
			ctx.LogInfof("[Database] Save file %s is empty, ignoring", latestBackupFile)
			return nil
		}

		err = json.Unmarshal(content, &data)
		if err != nil {
			ctx.LogErrorf("[Database] Error unmarshalling backup file %s: %v", latestBackupFile, err)
			return err
		}
		j.dataMutex.Lock()
		j.data = data
		j.dataMutex.Unlock()
		ctx.LogInfof("[Database] Successfully recovered data from backup file %s", latestBackupFile)
		return nil
	}

	return nil
}

func (j *JsonDatabase) loadFromFile(ctx basecontext.ApiContext) error {
	var data Data
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

	j.dataMutex.Lock()
	j.data = data
	j.connected = true
	j.dataMutex.Unlock()
	return nil
}

func (j *JsonDatabase) loadFromEmpty(ctx basecontext.ApiContext) error {
	ctx.LogInfof("[Database] Database file is empty, creating new file")
	j.dataMutex.Lock()
	j.data = Data{
		Users:            make([]models.User, 0),
		Claims:           make([]models.Claim, 0),
		Roles:            make([]models.Role, 0),
		ApiKeys:          make([]models.ApiKey, 0),
		PackerTemplates:  make([]models.PackerTemplate, 0),
		ManifestsCatalog: make([]models.CatalogManifest, 0),
	}
	j.dataMutex.Unlock()

	if j.Config.AutoRecover {
		// Check if there are any backup files available
		if err := j.recoverFromBackupFile(ctx); err != nil {
			ctx.LogErrorf("[Database] Error recovering from backup file: %v", err)
			return err
		}
	}

	if err := j.SaveNow(ctx); err != nil {
		ctx.LogErrorf("[Database] Error saving database file: %v", err)
		return err
	}

	j.connected = true
	return nil
}

func (j *JsonDatabase) IsDataFileEmpty(ctx basecontext.ApiContext) (bool, error) {
	// Retry opening the file every 200ms for 10 times
	retryCount := 0
	maxRetries := 10
	retryInterval := 200 * time.Millisecond

	// Adding a delay to allow for slow mounts to be ready
	for {
		if _, err := os.Stat(j.filename); os.IsNotExist(err) {
			ctx.LogInfof("[Database] Database file does not exist, creating new file")

			if retryCount >= maxRetries {
				ctx.LogErrorf("[Database] Error opening database file after %d retries: %v", maxRetries, err)
				break
			}
			retryCount++
			time.Sleep(retryInterval)
		} else {
			break
		}
	}

	if _, err := os.Stat(j.filename); os.IsNotExist(err) {
		ctx.LogInfof("[Database] Database file does not exist, creating new file")
		file, err := os.Create(j.filename)
		if err != nil {
			ctx.LogErrorf("[Database] Error creating database file: %v", err)
			return true, err
		}
		if err := file.Close(); err != nil {
			return true, err
		}
	}

	file, err := os.Open(j.filename)
	if err != nil {
		ctx.LogErrorf("[Database] Error opening database file: %v", err)
		return true, err
	}

	defer file.Close()

	isEmpty := false
	file.Close()

	fileContent, _ := helper.ReadFromFile(j.filename)
	isEmpty = len(fileContent) == 0

	return isEmpty, nil
}

func (j *JsonDatabase) removeGlobFiles(ctx basecontext.ApiContext, glob string) {
	filesToDelete, err := filepath.Glob(j.filename + glob)
	if err != nil {
		ctx.LogErrorf("[Database] Error finding files to delete: %v", err)
		return
	}

	for _, file := range filesToDelete {
		if err := os.Remove(file); err != nil {
			ctx.LogErrorf("[Database] Error deleting file %s: %v", file, err)
		}
	}
}
