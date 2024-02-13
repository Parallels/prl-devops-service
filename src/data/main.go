package data

import (
	"encoding/json"
	"os"
	"sync"

	"github.com/Parallels/pd-api-service/basecontext"
	"github.com/Parallels/pd-api-service/config"
	"github.com/Parallels/pd-api-service/data/models"
	"github.com/Parallels/pd-api-service/errors"
	"github.com/Parallels/pd-api-service/security"

	"github.com/cjlapao/common-go/helper"
)

var (
	ErrDatabaseNotConnected = errors.NewWithCode("the database is not connected", 500)
	ErrNotAuthorized        = errors.NewWithCode("not authorized to view record", 403)
)

var memoryDatabase *JsonDatabase

type Data struct {
	Users             []models.User             `json:"users"`
	Claims            []models.Claim            `json:"claims"`
	Roles             []models.Role             `json:"roles"`
	ApiKeys           []models.ApiKey           `json:"api_keys"`
	PackerTemplates   []models.PackerTemplate   `json:"virtual_machine_templates"`
	ManifestsCatalog  []models.CatalogManifest  `json:"catalog_manifests"`
	OrchestratorHosts []models.OrchestratorHost `json:"orchestrator_hosts"`
}

type JsonDatabase struct {
	connected   bool
	isSaving    bool
	saveProcess chan bool
	filename    string
	saveMutex   sync.Mutex
	saveQueue   []saveRequest
	data        Data
}

func NewJsonDatabase(filename string) *JsonDatabase {
	if memoryDatabase != nil {
		return memoryDatabase
	}

	memoryDatabase = &JsonDatabase{
		connected:   false,
		isSaving:    false,
		filename:    filename,
		saveProcess: make(chan bool),
		data:        Data{},
	}

	rootContext := basecontext.NewRootBaseContext()
	go memoryDatabase.ProcessSaveQueue(rootContext)
	_ = memoryDatabase.Load(rootContext)

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

	fileInfo, err := os.Stat(j.filename)
	if err != nil {
		ctx.LogErrorf("[Database] Error getting database file info: %v", err)
		return err
	}

	if fileInfo.Size() == 0 {
		ctx.LogInfof("[Database] Database file is empty, creating new file")
		j.data = Data{
			Users:            make([]models.User, 0),
			Claims:           make([]models.Claim, 0),
			Roles:            make([]models.Role, 0),
			ApiKeys:          make([]models.ApiKey, 0),
			PackerTemplates:  make([]models.PackerTemplate, 0),
			ManifestsCatalog: make([]models.CatalogManifest, 0),
		}

		err = j.Save(ctx)
		if err != nil {
			ctx.LogErrorf("[Database] Error saving database file: %v", err)
			return err
		}

		j.connected = true
		return nil
	} else {
		ctx.LogInfof("[Database] Database file is not empty, loading data")

		// Backup the file before attempting to read it
		backupFilename := j.filename + ".bak"
		err := helper.CopyFile(j.filename, backupFilename)
		if err != nil {
			ctx.LogErrorf("[Database] Error creating backup file: %v", err)
			return err
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

type saveRequest struct {
	ctx basecontext.ApiContext
	wg  *sync.WaitGroup
}

func (j *JsonDatabase) Save(ctx basecontext.ApiContext) error {
	wg := &sync.WaitGroup{}
	wg.Add(1)

	ctx.LogDebugf("[Database] Enqueuing save request")
	saveRequest := saveRequest{
		ctx: ctx,
		wg:  wg,
	}

	j.saveQueue = append(j.saveQueue, saveRequest)

	j.saveProcess <- true
	wg.Wait()
	return nil
}

func (j *JsonDatabase) ProcessSaveQueue(ctx basecontext.ApiContext) {
	for {
		<-j.saveProcess
		for len(j.saveQueue) > 0 {
			request := j.saveQueue[0]
			j.saveQueue = j.saveQueue[1:]
			if err := j.processSave(ctx); err != nil {
				ctx.LogErrorf("[Database] Error saving database: %v", err)
			}
			request.wg.Done()
		}
	}
}

func (j *JsonDatabase) processSave(ctx basecontext.ApiContext) error {
	j.saveMutex.Lock()

	cfg := config.Get()
	// Backup the file before attempting to read it
	backupFilename := j.filename + ".save.bak"
	err := helper.CopyFile(j.filename, backupFilename)
	if err != nil {
		ctx.LogErrorf("[Database] Error creating backup file: %v", err)
		return err
	}

	ctx.LogDebugf("[Database] Saving database to %s", j.filename)
	j.isSaving = true
	if j.filename == "" {
		j.saveMutex.Unlock()
		return errors.NewWithCode("the database filename is not set", 500)
	}

	// Trying to open the file and waiting for it to be ready
	var file *os.File
	for {
		var fileOpenError error
		file, fileOpenError = os.OpenFile(j.filename, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0o600)
		if fileOpenError == nil {
			break
		}
	}

	defer file.Close()

	jsonString, err := json.MarshalIndent(j.data, "", "  ")
	if err != nil {
		j.isSaving = false
		j.saveMutex.Unlock()
		return errors.NewFromError(err)
	}

	if cfg.EncryptionPrivateKey() != "" {
		encJsonString, err := security.EncryptString(cfg.EncryptionPrivateKey(), string(jsonString))
		if err != nil {
			_, saveErr := file.Write(jsonString)
			if saveErr != nil {
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

	_, err = file.Write(jsonString)
	if err != nil {
		j.isSaving = false
		j.saveMutex.Unlock()
		return err
	}

	if err := file.Close(); err != nil {
		j.isSaving = false
		j.saveMutex.Unlock()
		return err
	}

	j.isSaving = false
	j.saveMutex.Unlock()
	return nil
}
