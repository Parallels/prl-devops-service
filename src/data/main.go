package data

import (
	"encoding/json"
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
	Schema            models.DatabaseSchema                `json:"schema"`
	Configuration     *models.Configuration                `json:"configuration"`
	Users             []models.User                        `json:"users"`
	Claims            []models.Claim                       `json:"claims"`
	Roles             []models.Role                        `json:"roles"`
	ApiKeys           []models.ApiKey                      `json:"api_keys"`
	PackerTemplates   []models.PackerTemplate              `json:"virtual_machine_templates"`
	ManifestsCatalog  []models.CatalogManifest             `json:"catalog_manifests"`
	OrchestratorHosts []models.OrchestratorHost            `json:"orchestrator_hosts"`
	HostsVMSnapshots  []models.HostsVMSnapshotsRecord      `json:"orchestrator_snapshots"`
	ReverseProxy      *models.ReverseProxy                 `json:"reverse_proxy"`
	ReverseProxyHosts []models.ReverseProxyHost            `json:"reverse_proxy_hosts"`
	CatalogManagers   []models.CatalogManager              `json:"catalog_managers"`
	Jobs              []models.Job                         `json:"jobs"`
	VMSnapshots       []models.VMSnapshots                 `json:"vm_snapshots"`
	EnrollmentTokens  []models.OrchestratorEnrollmentToken `json:"enrollment_tokens"`
	UserConfigs       []models.UserConfig                  `json:"user_configs"`
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

	// Start ghost job cleanup goroutine to detect and cancel stalled jobs
	go func() {
		ticker := time.NewTicker(time.Duration(constants.GhostJobCheckIntervalSeconds) * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-memoryDatabase.cancel:
				ctx.LogInfof("[Database] Ghost job cleanup stopped")
				return
			case <-ticker.C:
				memoryDatabase.DetectStaleJobs(ctx)
			}
		}
	}()

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
	defer j.saveMutex.Unlock()

	cfg := config.Get()
	if cfg.GetRunningCommand() != constants.API_COMMAND && cfg.GetRunningCommand() != "" {
		ctx.LogDebugf("[Database] Skipping save request, command running: %s", cfg.GetRunningCommand())
		return nil
	}

	if j.filename == "" {
		return errors.NewWithCode("the database filename is not set", 500)
	}

	ctx.LogDebugf("[Database] Saving database to %s", j.filename)
	j.isSaving = true
	defer func() { j.isSaving = false }()

	// Acquire a cross-process exclusive lock so that multiple service instances
	// pointing at the same database directory cannot corrupt each other's saves.
	lock, err := acquireFileLock(j.filename + ".lock")
	if err != nil {
		ctx.LogDebugf("[Database] Error acquiring database lock: %v", err)
		return errors.NewFromError(err)
	}
	defer lock.release()

	// Marshal the data while holding only the read lock.
	j.dataMutex.RLock()
	jsonString, err := json.MarshalIndent(j.data, "", "  ")
	j.dataMutex.RUnlock()
	if err != nil {
		ctx.LogDebugf("[Database] Error marshalling data: %v", err)
		return errors.NewFromError(err)
	}
	ctx.LogDebugf("[Database] Data marshalled successfully")

	// Encrypt the data before saving it, if a key is configured.
	if cfg.EncryptionPrivateKey() != "" {
		encJsonString, encErr := security.EncryptString(cfg.EncryptionPrivateKey(), string(jsonString))
		if encErr != nil {
			ctx.LogDebugf("[Database] Error encrypting data: %v", encErr)
			return errors.NewFromError(encErr)
		}
		jsonString = encJsonString
	}

	// Atomic save: write to a uniquely-named temp file in the SAME directory,
	// flush it to stable storage, then rename it over the destination. POSIX
	// rename(2) on the same filesystem is atomic, so a crash, kill, or power loss
	// can never leave the database missing or half-written. We never delete the
	// live database before the new one is in place.
	dir := filepath.Dir(j.filename)
	tempFile, err := os.CreateTemp(dir, filepath.Base(j.filename)+".*.save")
	if err != nil {
		ctx.LogDebugf("[Database] Error creating temp file: %v", err)
		return errors.NewFromError(err)
	}
	tempFileName := tempFile.Name()

	// Best-effort cleanup of the temp file if we fail before the rename succeeds.
	renamed := false
	defer func() {
		if !renamed {
			_ = os.Remove(tempFileName)
		}
	}()

	ctx.LogDebugf("[Database] Writing data to temp file %s", tempFileName)
	if _, err = tempFile.Write(jsonString); err != nil {
		ctx.LogDebugf("[Database] Error writing data to temp file: %v", err)
		_ = tempFile.Close()
		return errors.NewFromError(err)
	}

	// Flush to disk before swapping the file into place so the rename can never
	// expose a partially-written file after a crash.
	if err = tempFile.Sync(); err != nil {
		ctx.LogDebugf("[Database] Error syncing temp file: %v", err)
		_ = tempFile.Close()
		return errors.NewFromError(err)
	}

	if err = tempFile.Close(); err != nil {
		ctx.LogDebugf("[Database] Error closing temp file: %v", err)
		return errors.NewFromError(err)
	}

	if err = os.Rename(tempFileName, j.filename); err != nil {
		ctx.LogDebugf("[Database] Error renaming temp file into place: %v", err)
		return errors.NewFromError(err)
	}
	renamed = true

	ctx.LogDebugf("[Database] File %s saved successfully", j.filename)
	return nil
}

// fileLock holds an OS-level advisory lock on a lock file for the lifetime of a
// save, preventing concurrent saves from separate processes that share the same
// database directory.
type fileLock struct {
	f *os.File
}

// acquireFileLock opens (creating if needed) the given lock file and blocks
// until it holds an exclusive OS-level lock on it.
func acquireFileLock(path string) (*fileLock, error) {
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE, 0o600)
	if err != nil {
		return nil, err
	}
	if err := lockFileDescriptor(f); err != nil {
		_ = f.Close()
		return nil, err
	}
	return &fileLock{f: f}, nil
}

// release unlocks and closes the underlying lock file. Safe to call on a nil lock.
func (l *fileLock) release() {
	if l == nil || l.f == nil {
		return
	}
	_ = unlockFileDescriptor(l.f)
	_ = l.f.Close()
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
	if user := ctx.GetUser(); user != nil {
		dbRecord.LockedBy = user.Email
	}
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
		err = json.Unmarshal(content, &data)
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

	// Handle recovery of ongoing jobs
	j.RecoverOngoingJobs(ctx)

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
		HostsVMSnapshots: make([]models.HostsVMSnapshotsRecord, 0),
		VMSnapshots:      make([]models.VMSnapshots, 0),
		CatalogManagers:  make([]models.CatalogManager, 0),
		Jobs:             make([]models.Job, 0),
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
		_, err := os.Stat(j.filename)
		if err == nil {
			break
		}

		if os.IsNotExist(err) {
			// If it definitely doesn't exist, don't wait 10 times for a mount
			break
		}

		ctx.LogInfof("[Database] Database file not accessible yet (err: %v), waiting for volume mount...", err)

		if retryCount >= maxRetries {
			ctx.LogErrorf("[Database] Error opening database file after %d retries: %v", maxRetries, err)
			break
		}
		retryCount++
		time.Sleep(retryInterval)
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
