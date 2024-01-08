package pdfile

type PDFile struct {
	raw            []string              `json:"-" yaml:"-"`
	Insecure       bool                  `json:"-" yaml:"-"`
	From           string                `json:"FROM,omitempty" yaml:"FROM,omitempty"`
	Authentication *PDFileAuthentication `json:"AUTHENTICATION,omitempty" yaml:"AUTHENTICATION,omitempty"`
	Description    string                `json:"DESCRIPTION,omitempty" yaml:"DESCRIPTION,omitempty"`
	CatalogId      string                `json:"CATALOG_ID,omitempty" yaml:"CATALOG_ID,omitempty"`
	Version        string                `json:"VERSION,omitempty" yaml:"VERSION,omitempty"`
	Architecture   string                `json:"ARCHITECTURE,omitempty" yaml:"ARCHITECTURE,omitempty"`
	LocalPath      string                `json:"LOCAL_PATH,omitempty" yaml:"LOCAL_PATH,omitempty"`
	Destination    string                `json:"DESTINATION,omitempty" yaml:"DESTINATION,omitempty"`
	MachineName    string                `json:"MACHINE_NAME,omitempty" yaml:"MACHINE_NAME,omitempty"`
	Owner          string                `json:"OWNER,omitempty" yaml:"OWNER,omitempty"`
	StartAfterPull bool                  `json:"START_AFTER_PULL,omitempty" yaml:"START_AFTER_PULL,omitempty"`
	Roles          []string              `json:"ROLES,omitempty" yaml:"ROLES,omitempty"`
	Claims         []string              `json:"CLAIMS,omitempty" yaml:"CLAIMS,omitempty"`
	Tags           []string              `json:"TAGS,omitempty" yaml:"TAGS,omitempty"`
	Provider       *PDFileProvider       `json:"PROVIDER,omitempty" yaml:"PROVIDER,omitempty"`
	Command        string                `json:"COMMAND,omitempty" yaml:"COMMAND,omitempty"`
	Execute        []string              `json:"EXECUTE,omitempty" yaml:"EXECUTE,omitempty"`
	Operation      string                `json:"RUN,omitempty" yaml:"RUN,omitempty"`
}

type PDFileAuthentication struct {
	Username string `json:"USERNAME,omitempty" yaml:"USERNAME,omitempty"`
	Password string `json:"PASSWORD,omitempty" yaml:"PASSWORD,omitempty"`
	ApiKey   string `json:"API_KEY,omitempty" yaml:"API_KEY,omitempty"`
}

type PDFileProvider struct {
	Name       string            `json:"NAME,omitempty" yaml:"NAME,omitempty"`
	Attributes map[string]string `json:"ATTRIBUTES,omitempty" yaml:"ATTRIBUTES,omitempty"`
}

func NewPdFile() *PDFile {
	return &PDFile{
		raw:    []string{},
		Roles:  []string{},
		Claims: []string{},
	}
}

type PullResponse struct {
	MachineId    string `json:"machine_id,omitempty" yaml:"machine_id,omitempty"`
	MachineName  string `json:"machine_name,omitempty" yaml:"machine_name,omitempty"`
	CatalogId    string `json:"catalog_id,omitempty" yaml:"catalog_id,omitempty"`
	Version      string `json:"version,omitempty" yaml:"version,omitempty"`
	Architecture string `json:"architecture,omitempty" yaml:"architecture,omitempty"`
	Type         string `json:"type,omitempty" yaml:"type,omitempty"`
}
