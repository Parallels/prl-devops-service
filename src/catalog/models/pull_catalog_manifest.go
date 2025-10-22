package models

import (
	"fmt"

	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/catalog/cleanupservice"
	"github.com/Parallels/prl-devops-service/config"
	"github.com/Parallels/prl-devops-service/constants"
	"github.com/Parallels/prl-devops-service/errors"
	"github.com/Parallels/prl-devops-service/serviceprovider"
	"github.com/Parallels/prl-devops-service/serviceprovider/system"
)

var (
	ErrPullMissingPath        = errors.NewWithCode("missing path", 400)
	ErrPullMissingCatalogId   = errors.NewWithCode("missing catalog id", 400)
	ErrPullMissingMachineName = errors.NewWithCode("missing machine name", 400)
	ErrMissingConnection      = errors.NewWithCode("missing connection", 400)
)

type PullCatalogManifestRequest struct {
	architecture       string
	CatalogId          string            `json:"catalog_id"`
	Version            string            `json:"version,omitempty"`
	Architecture       string            `json:"architecture,omitempty"`
	Owner              string            `json:"owner,omitempty"`
	MachineName        string            `json:"machine_name,omitempty"`
	Path               string            `json:"path,omitempty"`
	Connection         string            `json:"connection,omitempty"`
	ProviderMetadata   map[string]string `json:"provider_metadata,omitempty"`
	StartAfterPull     bool              `json:"start_after_pull,omitempty"`
	LocalMachineFolder string            `json:"-"`
	ProgressChannel    chan int          `json:"-"`
	FileNameChannel    chan string       `json:"-"`
	StepChannel        chan string       `json:"-"`
	FromPdf            bool              `json:"-"`
	AmplitudeEvent     string            `json:"client,omitempty"`
}

func (r *PullCatalogManifestRequest) Validate() error {
	if r.CatalogId == "" {
		return ErrPullMissingCatalogId
	}
	if r.Version == "" {
		r.Version = constants.LATEST_TAG
	}
	if r.MachineName == "" {
		return ErrPullMissingMachineName
	}
	if r.Connection == "" {
		return ErrMissingConnection
	}

	ctx := basecontext.NewRootBaseContext()
	svcCtl := system.Get()
	cfg := config.Get()
	arch, err := svcCtl.GetArchitecture(ctx)
	if err != nil {
		return errors.New("unable to determine architecture")
	}
	r.architecture = arch

	if r.Owner == "" {
		r.Owner = cfg.GetKey(constants.CURRENT_USER_ENV_VAR)
	}

	if r.Path == "" {
		prl := serviceprovider.Get().ParallelsDesktopService
		if prl == nil {
			return errors.New("Local Path is required and we are unable to determine it without Parallels Desktop Service")
		}
		userPath, err := prl.GetUserHome(ctx, r.Owner)
		if err != nil {
			return fmt.Errorf("unable to determine user %v home for path", r.Owner)
		}
		r.Path = userPath
	}

	return nil
}

type PullCatalogManifestResponse struct {
	ID             string                         `json:"id"`
	MachineID      string                         `json:"machine_id,omitempty"`
	CatalogId      string                         `json:"catalog_id,omitempty"`
	Version        string                         `json:"version,omitempty"`
	Architecture   string                         `json:"architecture,omitempty"`
	LocalPath      string                         `json:"local_path,omitempty"`
	MachineName    string                         `json:"machine_name,omitempty"`
	Manifest       *VirtualMachineCatalogManifest `json:"manifest,omitempty"`
	LocalCachePath string                         `json:"local_cache_path,omitempty"`
	CleanupRequest *cleanupservice.CleanupService `json:"-"`
	Errors         []error                        `json:"-"`
}

func NewPullCatalogManifestResponse() *PullCatalogManifestResponse {
	return &PullCatalogManifestResponse{
		CleanupRequest: cleanupservice.NewCleanupService(),
		Errors:         []error{},
	}
}

func (m *PullCatalogManifestResponse) HasErrors() bool {
	return len(m.Errors) > 0
}

func (m *PullCatalogManifestResponse) AddError(err error) {
	m.Errors = append(m.Errors, err)
}
